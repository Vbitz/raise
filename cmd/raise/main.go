// raise is the Raise worker agent. It connects to the server using a long-lived Websocket connection.
package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"path"

	"github.com/Vbitz/raise/v2/pkg/worker"
)

var (
	serverAddress     = flag.String("server", "", "The address of the server to connect to.")
	serverCertificate = flag.String("cert", "", "The certificate of the server to connect to.")
)

type ConfigFile struct {
	ServerAddress         string
	ServerCertificate     string
	ClientCertificatePath string
	ClientKeyPath         string
}

func loadConfig() error {
	exec, err := os.Executable()
	if err != nil {
		return err
	}

	execDir := path.Dir(exec)

	configPath := path.Join(execDir, "worker.json")

	configContent, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	var config ConfigFile

	err = json.Unmarshal(configContent, &config)
	if err != nil {
		return err
	}

	// Set config values.
	*serverAddress = config.ServerAddress
	*serverCertificate = config.ServerCertificate

	return nil
}

func main() {
	flag.Parse()

	err := loadConfig()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	worker := worker.NewWorker(*serverAddress, *serverCertificate)

	err = worker.Connect()
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
}
