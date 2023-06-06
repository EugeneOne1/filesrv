# File Server

This is a simple file server based on [http.FileServer] Go standard library with
custom front-end for directories.  It is not a security-aware tool, so avoid
using it within public Internet.

## Running

To run the server you may simply use the `go run` command:

```sh
cd filesrv && go run ./cmd/srv.go
```

Alternatively, you may build the binary and run it:

```sh
cd filesrv && go build ./cmd/srv.go && ./srv
```

These scenarios are also covered by the `Makefile`.  See the `Makefile` itself
for more details.

[http.FileServer](https://pkg.go.dev/net/http#FileServer)
