package cmd

import (
	"fmt"
	"net"
	"net/url"
	"os"
)

// printListenAddrs prints the addresses the server is listening on appending
// the specified port to each one.
func printListenAddrs(port string) (err error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return fmt.Errorf("getting interface addresses: %w", err)
	} else if len(addrs) == 0 {
		return fmt.Errorf("no interface addresses found")
	}

	fmt.Println("Listening on:")
	hn, err := os.Hostname()
	if err == nil {
		fmt.Printf("\t%s\n", &url.URL{
			Scheme: "http",
			Host:   net.JoinHostPort(hn, port),
		})
	}

	for _, addr := range addrs {
		if n, ok := addr.(*net.IPNet); !ok {
			continue
		} else {
			fmt.Printf("\t%s\n", &url.URL{
				Scheme: "http",
				Host:   net.JoinHostPort(n.IP.String(), port),
			})
		}
	}

	return nil
}
