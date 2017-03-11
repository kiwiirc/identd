# kiwiirc - identd
An RPC controlled identd server supporting:
* Multiple network interfaces
* RPC for applications to add/remove identd entries

## Usage
Specify the interface for the identd server:
> ./identd -identd=tcp://0.0.0.0:113

Specify the interface for the RPC socket:
> ./identd -rpc=tcp://127.0.0.1:11333

> ./identd -rpc=unix:///tmp/identd.sock


## RPC Commands
All commands are sent by connecting to the RPC interface, sending the command line followed by a new line.

* `add <username> <lport> <rport> [interface]`
* `del <lport> <rport> [interface]`
* `lookup <lport> <rport> [interface]`
* `id <app id>`
* `clear`

### add
`add <username> <local port> <remote port> [local interface address]`

`add someuser 3293 6667 1.1.1.1`

Adds an identd entry. The optional local interface address defaults to 0.0.0.0.

### del
`del <local port> <remote port> [local interface address]`

`del 3293 6667 1.1.1.1`

Delete an identd entry. The optional local interface address defaults to 0.0.0.0.

### lookup
`lookup <local port> <remote port> [local interface address]`

`lookup 3293 6667 1.1.1.1`

Looks up an identd entry. Returns a period "." as the username if not found. The optional local interface address defaults to 0.0.0.0.

### id
`id <app id>`

`id myapp`

Sets the application ID for this RPC connection. It may be any string excluding whitespace. If there are several applications using the identd you may need to flush out all identd entries for your application only (eg. on application restart). See the `clear` command.

### clear
`clear`

Flushes out all identd entries for the current application. The current
application is identified by the string sent with the `id` command.
