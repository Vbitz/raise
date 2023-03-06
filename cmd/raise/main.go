// raise is the Raise worker agent. It connects to the server using a long-lived Websocket connection.
package main

import "flag"

var (
	serverAddress     = flag.String("server", "", "The address of the server to connect to.")
	serverCertificate = flag.String("cert", "", "The certificate of the server to connect to.")
)

func main() {
	flag.Parse()
}
