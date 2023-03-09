// ra is the client frontend for Raise.
// ra is installed and authenticated on clients granting access to any workers.
// Scripting is provided though a Starlark API. ra takes a single argument as a argument.
package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"flag"
	"log"
	"os"
	"path"

	"github.com/Vbitz/raise/v2/pkg/client"
	"github.com/Vbitz/raise/v2/pkg/common"
	"github.com/Vbitz/raise/v2/pkg/security"
	"github.com/Vbitz/raise/v2/pkg/star"
)

var (
	serverAddress     = flag.String("serverAddress", "", "The address of the server to connect to.")
	serverCertificate = flag.String("serverCertificate", "", "The certificate of the server to connect to. The certificate is base64 encoded in DER format.")
	clientCertificate = flag.String("clientCertificate", "", "The certificate the client uses to authenticate to the server.")
	clientKey         = flag.String("clientKey", "", "The private key the client uses to authenticate to the server.")
	version           = flag.Bool("version", false, "Print the current version and exit.")
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

	configPath := path.Join(execDir, "client.json")

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
	*clientCertificate = config.ClientCertificatePath
	*clientKey = config.ClientKeyPath

	return nil
}

func writeCertAndKey(certBytes []byte, privBytes []byte) error {
	exec, err := os.Executable()
	if err != nil {
		return err
	}

	execDir := path.Dir(exec)

	err = os.WriteFile(path.Join(execDir, "client.crt"), certBytes, os.ModePerm)
	if err != nil {
		return err
	}

	err = os.WriteFile(path.Join(execDir, "client.key"), privBytes, os.ModePerm)
	if err != nil {
		return err
	}

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

	filename := flag.Arg(0)
	if filename == "" {
		log.Fatalf("no script specified")
	}

	fileContents, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("error reading script: %v", err)
	}

	if *clientCertificate == "" || *clientKey == "" {
		// Generate a new certificate and key then exit.
		log.Printf("No certificate or key specified. Generating a keypair now.")

		cert, certBytes, err := security.GenerateCertificatePair()
		if err != nil {
			log.Fatalf("failed to generate certificate: %v", err)
		}

		privBytes := x509.MarshalPKCS1PrivateKey(cert.PrivateKey.(*rsa.PrivateKey))

		err = writeCertAndKey(certBytes, privBytes)
		if err != nil {
			log.Fatalf("failed to write certificate and key: %v", err)
		}

		return
	}

	client := client.NewClient(
		*serverAddress,
		*serverCertificate,
		*clientCertificate,
		*clientKey,
	)
	defer client.Close()

	engine := star.NewEngine()

	err = engine.RunFile(client, nil, filename, fileContents)
	if err != nil {
		log.Fatalf("error running script: %v", err)
	}
}
