package worker

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/Vbitz/raise/v2/pkg/proto"
	"github.com/cenkalti/rpc2"
	"github.com/gobwas/ws"
)

type Worker struct {
	serverCertificate string
	serverAddress     string
	name              string

	rpcClient *rpc2.Client
}

// GetInfo implements proto.WorkerService
func (w *Worker) GetInfo(client *rpc2.Client, req proto.GetInfoReq, resp *proto.GetInfoResp) error {
	var err error

	*resp = proto.GetInfoResp{}

	resp.Hostname, err = os.Hostname()
	if err != nil {
		return err
	}

	resp.HomeDir, err = os.UserHomeDir()
	if err != nil {
		return err
	}

	resp.OperatingSystem = runtime.GOOS
	resp.Architecture = runtime.GOARCH

	return nil
}

// SendMessage implements proto.WorkerService
func (w *Worker) SendMessage(client *rpc2.Client, req proto.SendMessageReq, resp *proto.SendMessageResp) error {
	*resp = proto.SendMessageResp{}

	if req.Kind == proto.MessageReadFile {
		content, err := os.ReadFile(req.Filename)
		if err != nil {
			return err
		}

		resp.Content = content

		return nil
	} else if req.Kind == proto.MessageWriteFile {
		err := os.WriteFile(req.Filename, req.Content, os.ModePerm)
		if err != nil {
			return err
		}

		return nil
	} else if req.Kind == proto.MessageRunScript {
		content, err := w.RunScript(string(req.Content))
		if err != nil {
			return err
		}

		resp.Content = content

		return nil
	} else {
		return fmt.Errorf("unknown message kind: %s", req.Kind)
	}
}

// Ping implements proto.WorkerService
func (w *Worker) Ping(client *rpc2.Client, req proto.PingReq, resp *proto.PingResp) error {
	log.Printf("Got ping request")

	*resp = proto.PingResp{
		Message: fmt.Sprintf("Hello from worker %s", w.name),
	}
	return nil
}

func (w *Worker) RunScript(script string) ([]byte, error) {
	var cmd *exec.Cmd

	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		cmd = exec.Command("/bin/bash", "-s")
		cmd.Stdin = bytes.NewReader([]byte(script))
	} else if runtime.GOOS == "windows" {
		cmd = exec.Command("powershell.exe", "-Command", "-")
		cmd.Stdin = bytes.NewReader([]byte(script))
	} else {
		return nil, fmt.Errorf("operating system %s not supported", runtime.GOOS)
	}

	stdoutBuffer := new(bytes.Buffer)

	cmd.Stdout = stdoutBuffer
	cmd.Stderr = stdoutBuffer

	err := cmd.Start()
	if err != nil {
		return nil, err
	}

	err = cmd.Wait()
	if err != nil {
		return nil, err
	}

	return stdoutBuffer.Bytes(), nil
}

func (w *Worker) Connect() error {
	dialer := ws.Dialer{}

	certPool := x509.NewCertPool()

	pemBytes, err := base64.StdEncoding.DecodeString(w.serverCertificate)
	if err != nil {
		return err
	}

	ok := certPool.AppendCertsFromPEM(pemBytes)
	if !ok {
		return fmt.Errorf("failed to add server certificate")
	}

	dialer.TLSConfig = &tls.Config{}

	dialer.TLSConfig.ServerName = "localhost"

	dialer.TLSConfig.RootCAs = certPool

	conn, _, _, err := dialer.Dial(context.Background(), w.serverAddress+"/worker")
	if err != nil {
		return err
	}

	w.rpcClient = rpc2.NewClient(conn)

	go w.rpcClient.Run()

	w.rpcClient.Handle(proto.Common_Ping, w.Ping)
	w.rpcClient.Handle(proto.Common_SendMessage, w.SendMessage)
	w.rpcClient.Handle(proto.Common_GetInfo, w.GetInfo)

	// Send the hello message to register the worker with the server.
	var helloResp proto.HelloResp
	err = w.rpcClient.Call(proto.Control_Hello, proto.HelloReq{
		Name: w.name,
	}, &helloResp)
	if err != nil {
		return err
	}

	log.Printf("worker registered as %s on server %s", w.name, w.serverAddress)

	for {
		time.Sleep(1 * time.Hour)
	}
}

var (
	_ proto.WorkerService = &Worker{}
)

func NewWorker(serverAddress string, serverCertificate string, name string) *Worker {
	return &Worker{
		serverCertificate: serverCertificate,
		serverAddress:     serverAddress,
		name:              name,
	}
}
