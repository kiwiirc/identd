package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
)

var identLookup *IdentdLookup
var isLogging bool

func main() {
	rpcListenerStr := flag.String("rpc", "tcp://:1133", "RPC socket listener")
	identdListenerStr := flag.String("identd", "tcp://:113", "Identd listener")
	logging := flag.Bool("v", false, "Verbose logging output")
	flag.Parse()

	isLogging = *logging

	rpcListener, err := listenerFromString(*rpcListenerStr)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Printf("RPC listening on %s", rpcListener.Addr().String())
	}

	identdListener, err := listenerFromString(*identdListenerStr)
	if err != nil {
		log.Fatal(err)
	}

	identLookup = MakeIdentdLookup()

	go listenForRpcSockets(rpcListener)
	go listenForIdentdSockets(identdListener)

	c := make(chan bool)
	<-c
}

func listenerFromString(inp string) (net.Listener, error) {
	parts := strings.Split(inp, "://")
	if len(parts) != 2 {
		return nil, errors.New("Invalid interface, " + inp)
	}

	serv, err := net.Listen(parts[0], parts[1])
	return serv, err
}

func listenForRpcSockets(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err.Error())
			continue
		}

		go rpcSocketHandler(conn)
	}
}

func rpcSocketHandler(conn net.Conn) {
	// The default app ID if one isn't set
	appID := "1"

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		if isLogging {
			log.Printf("Raw RPC in: %s", line)
		}
		parts := strings.Split(line, " ")

		if parts[0] == "id" {
			// id new_app_id
			if len(parts) > 1 {
				appID = parts[1]
			}

		} else if parts[0] == "add" {
			// add username lport rport [interface]
			if len(parts) < 4 {
				continue
			}

			username := parts[1]
			lport, _ := strconv.Atoi(parts[2])
			rport, _ := strconv.Atoi(parts[3])
			inet := "0.0.0.0"

			if len(parts) == 5 {
				inet = parts[4]
			}

			if lport > 0 && rport > 0 && username != "" {
				if isLogging {
					log.Printf("Adding entry %d %d %s %s %s", lport, rport, inet, username, appID)
				}
				identLookup.AddEntry(lport, rport, inet, username, appID)
			}

		} else if parts[0] == "del" {
			// del lport rport [interface]
			if len(parts) < 3 {
				continue
			}

			lport, _ := strconv.Atoi(parts[1])
			rport, _ := strconv.Atoi(parts[2])
			inet := "0.0.0.0"

			if len(parts) == 4 {
				inet = parts[3]
			}

			e := identLookup.Lookup(lport, rport, inet)

			if e != nil {
				if isLogging {
					log.Printf("Removing entry %d %d %s %s %s", e.LocalPort, e.RemotePort, e.Inet, e.Username, e.AppID)
				}
				identLookup.RemoveEntry(e)
			}

		} else if parts[0] == "clear" {
			// clear
			if isLogging {
				log.Printf("Clearing entries for app %s", appID)
			}
			identLookup.ClearAppID(appID)

		} else if parts[0] == "lookup" {
			// lookup lport rport [interface]
			if len(parts) < 3 {
				continue
			}

			lport, _ := strconv.Atoi(parts[1])
			rport, _ := strconv.Atoi(parts[2])
			inet := "0.0.0.0"

			if len(parts) == 4 {
				inet = parts[3]
			}

			e := identLookup.Lookup(lport, rport, inet)

			if e != nil {
				fmt.Fprintf(conn, "%d %d %s\r\n", e.LocalPort, e.RemotePort, e.Username)
			} else {
				fmt.Fprintf(conn, "%d %d .\r\n", lport, rport)
			}
		}
	}
}

func listenForIdentdSockets(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err.Error())
			continue
		}

		go identdSocketHandler(conn)
	}
}

func identdSocketHandler(conn net.Conn) {
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		if isLogging {
			log.Printf("Raw identd in: %s", line)
		}
		parts := strings.Split(line, ",")

		if len(parts) != 2 {
			conn.Close()
			continue
		}

		lport, _ := strconv.Atoi(strings.Trim(parts[0], " "))
		rport, _ := strconv.Atoi(strings.Trim(parts[1], " "))
		inet, _, _ := net.SplitHostPort(conn.LocalAddr().String())

		entry := identLookup.Lookup(lport, rport, inet)
		if entry == nil {
			if isLogging {
				log.Printf("Identd lookup for %d %d did not find a user", lport, rport)
			}
			fmt.Fprintf(conn, "%d, %d : ERROR : NO-USER", lport, rport)
		} else {
			if isLogging {
				log.Printf("Identd lookup for %d %d found %s", lport, rport, entry.Username)
			}
			fmt.Fprintf(conn, "%d, %d : USERID : KiwiIRC : %s", lport, rport, entry.Username)
		}

		conn.Close()
	}
}
