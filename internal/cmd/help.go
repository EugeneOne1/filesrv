package cmd

import (
	"fmt"
	"log"
	"net"
	"net/url"
	"os"

	qrcodeTerminal "github.com/Baozisoftware/qrcode-terminal-go"
)

// printListenAddrs prints the addresses the server is listening on appending
// the specified port to each one.  It also prints a QR code for the first
// found hostname.
func printListenAddrs(port string) (err error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return fmt.Errorf("getting interface addresses: %w", err)
	} else if len(addrs) == 0 {
		return fmt.Errorf("no interface addresses found")
	}

	fmt.Println("Available at:")

	qrPrinted := false

	for _, addr := range addrs {
		if n, ok := addr.(*net.IPNet); !ok {
			fmt.Printf("\t%s (not a net)\n", addr.String())

			continue
		} else {
			if !qrPrinted && n.IP.IsPrivate() {
				printQR((&url.URL{
					Scheme: "http",
					Host:   net.JoinHostPort(n.IP.String(), port),
				}).String())

				qrPrinted = true
			}

			fmt.Printf("\t%s\n", &url.URL{
				Scheme: "http",
				Host:   net.JoinHostPort(n.IP.String(), port),
			})
		}
	}

	if !qrPrinted {
		hn, err := os.Hostname()
		if err != nil {
			// This error is not critical?
			log.Printf("getting hostname: %s", err)

			return nil
		}

		printQR((&url.URL{
			Scheme: "http",
			Host:   net.JoinHostPort(hn, port),
		}).String())
	}

	return nil
}

func printQR(content string) {
	fmt.Println("Try also:")
	qrcodeTerminal.New2(
		qrcodeTerminal.ConsoleColors.BrightBlack,
		qrcodeTerminal.ConsoleColors.BrightWhite,
		qrcodeTerminal.QRCodeRecoveryLevels.Low,
	).Get(content).Print()
	fmt.Printf("\t%s\n", content)
}
