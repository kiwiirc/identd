package main

import "sync"

// Combine 2 ports into a single int. Eg. 65535, 65535 = 6553565535
func combinePorts(port1, port2 int) uint64 {
	return uint64((port1 * 100000) + port2)
}

// Split a combined int into 2 ports. Eg. 6553565535 = 65535, 65535
func splitPorts(inp uint64) (int, int) {
	port1 := inp / 100000
	port2 := inp - (port1 * 100000)
	return int(port1), int(port2)
}

type IdentdEntry struct {
	Key        uint64
	LocalPort  int
	RemotePort int
	Inet       string
	Username   string
	AppID      string
}

type IdentdLookup struct {
	sync.Mutex
	// Entries[inet][combined port] = IdentdEntry
	Entries map[string]map[uint64]*IdentdEntry
}

func MakeIdentdLookup() *IdentdLookup {
	return &IdentdLookup{Entries: make(map[string]map[uint64]*IdentdEntry)}
}

func (l *IdentdLookup) Lookup(localPort, remotePort int, inet string) *IdentdEntry {
	l.Lock()
	defer l.Unlock()

	inetEntries := l.Entries[inet]
	if inetEntries == nil {
		inetEntries, _ = l.Entries["0.0.0.0"]
	}

	if inetEntries == nil {
		return nil
	}

	key := combinePorts(localPort, remotePort)
	entry, _ := inetEntries[key]

	return entry
}

func (l *IdentdLookup) AddEntry(localPort, remotePort int, inet string, username string, appID string) *IdentdEntry {
	key := combinePorts(localPort, remotePort)
	entry := &IdentdEntry{
		Key:        key,
		LocalPort:  localPort,
		RemotePort: remotePort,
		Inet:       inet,
		Username:   username,
		AppID:      appID,
	}

	l.Lock()
	inetEntries := l.Entries[inet]
	if inetEntries == nil {
		l.Entries[inet] = make(map[uint64]*IdentdEntry)
		inetEntries = l.Entries[inet]
	}

	inetEntries[key] = entry
	l.Unlock()

	return entry
}

func (l *IdentdLookup) RemoveEntry(entry *IdentdEntry) {
	l.Lock()
	defer l.Unlock()

	inetEntries := l.Entries[entry.Inet]
	if inetEntries == nil {
		return
	}

	delete(inetEntries, entry.Key)
	if len(inetEntries) == 0 {
		delete(l.Entries, entry.Inet)
	}
}

func (l *IdentdLookup) ClearAppID(appID string) {
	l.Lock()
	defer l.Unlock()

	for _, inetEntries := range l.Entries {
		for key, entry := range inetEntries {
			if entry.AppID == appID {
				delete(inetEntries, key)
			}
		}
	}
}
