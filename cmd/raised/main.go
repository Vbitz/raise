// raised is the Raise control plane. It accepts commands from clients and forwards the commands to workers.
package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"github.com/Vbitz/raise/v2/pkg/server"
)

var (
	addr       = flag.String("addr", ":5634", "The address for the server to listen on.")
	certFile   = flag.String("cert", "", "The certificate file to use for HTTPS.")
	keyFile    = flag.String("key", "", "The key file to use for HTTPS.")
	clientList = flag.String("clientList", "", "A file containing a list of client keys to trust.")
)

func main() {
	flag.Parse()

	svr := server.NewServer(*addr, *certFile, *keyFile)

	clientListContent, err := os.ReadFile(*clientList)
	if err != nil {
		log.Fatalf("failed to read client list: %v", err)
	}

	for _, line := range strings.Split(string(clientListContent), "\n") {
		tokens := strings.Split(line, " ")

		client := server.Client{
			Name:              tokens[1],
			CertificateString: tokens[0],
		}

		err := svr.AddClient(client)
		if err != nil {
			log.Fatalf("failed to add client: %v", err)
		}
	}

	err = svr.Listen()
	if err != nil {
		log.Fatal(err)
	}
}
