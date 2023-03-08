package main

import (
	"encoding/base64"
	"flag"
	"log"
	"os"
	"time"

	"github.com/Vbitz/raise/v2/pkg/client"
	"github.com/Vbitz/raise/v2/pkg/server"
	"github.com/Vbitz/raise/v2/pkg/star"
	"github.com/Vbitz/raise/v2/pkg/worker"
)

var (
	serverAddr     = flag.String("addr", ":5634", "The address the server listens on.")
	serverCertFile = flag.String("serverCert", "testData/server.crt", "The certificate file to use for HTTPS.")
	serverKeyFile  = flag.String("serverKey", "testData/server.key", "The key file to use for HTTPS.")
	clientCertFile = flag.String("clientCert", "build/client.crt", "The certificate the client uses.")
	clientKeyFile  = flag.String("clientKey", "build/client.key", "The certificate the client uses.")
)

func main() {
	flag.Parse()

	filename := flag.Arg(0)
	if filename == "" {
		log.Fatalf("no script specified")
	}

	fileContents, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("error reading script: %v", err)
	}

	// Start the server.
	go func() {
		clientCertContent, err := os.ReadFile(*clientCertFile)
		if err != nil {
			log.Fatalf("failed to read client certificate: %v", err)
		}

		svr := server.NewServer(*serverAddr, *serverCertFile, *serverKeyFile)

		client := server.Client{
			Name:              "testingClient",
			CertificateString: base64.StdEncoding.EncodeToString(clientCertContent),
		}

		err = svr.AddClient(client)
		if err != nil {
			log.Fatalf("failed to add client: %v", err)
		}

		err = svr.Listen()
		if err != nil {
			log.Fatal(err)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	serverCertContent, err := os.ReadFile(*serverCertFile)
	if err != nil {
		log.Fatalf("Failed to read server certificate")
	}

	serverCert := base64.StdEncoding.EncodeToString(serverCertContent)

	// Start the worker.
	go func() {
		worker := worker.NewWorker("wss://localhost"+*serverAddr, serverCert, "testing")

		err := worker.Connect()
		if err != nil {
			log.Fatalf("worker failed to connect: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	// Start the client and have it execute the passed script.
	client := client.NewClient(
		"wss://localhost"+*serverAddr,
		serverCert,
		*clientCertFile,
		*clientKeyFile,
	)
	defer client.Close()

	engine := star.NewEngine()

	err = engine.RunFile(client, filename, fileContents)
	if err != nil {
		log.Fatalf("error running script: %v", err)
	}
}
