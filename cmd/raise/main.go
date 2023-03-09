// raise is the Raise worker agent. It connects to the server using a long-lived Websocket connection.
package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"path"
	"time"

	"github.com/Vbitz/raise/v2/pkg/common"
	"github.com/Vbitz/raise/v2/pkg/worker"
)

var (
	serverAddress     = flag.String("server", "", "The address of the server to connect to.")
	serverCertificate = flag.String("cert", "", "The certificate of the server to connect to.")
	name              = flag.String("name", "", "The name the worker identifies to the server.")
	version           = flag.Bool("version", false, "Print the current version and exit.")
)

type ConfigFile struct {
	ServerAddress     string
	ServerCertificate string
	WorkerName        string
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
	*name = config.WorkerName

	return nil
}

func main() {
	flag.Parse()

	if *version {
		log.Printf("%s", common.Commit)
	}

	err := loadConfig()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	worker := worker.NewWorker(*serverAddress, *serverCertificate, *name)

	for {
		log.Printf("attempting to connect to: %s", *serverAddress)
		err = worker.Connect()
		if err != nil {
			log.Printf("failed to connect: %v", err)
			time.Sleep(10 * time.Second)
		}
	}
}
