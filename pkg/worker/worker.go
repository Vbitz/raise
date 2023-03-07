package worker

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/Vbitz/raise/v2/pkg/proto"
	"github.com/cenkalti/rpc2"
	"github.com/gobwas/ws"
)

type Worker struct {
	serverCertificate string
	serverAddress     string

	rpcClient *rpc2.Client
}

// Ping implements proto.WorkerService
func (*Worker) Ping(client *rpc2.Client, req proto.PingReq, resp *proto.PingResp) error {
	*resp = proto.PingResp{}
	return nil
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

	dialer.TLSConfig.RootCAs = certPool

	conn, _, _, err := dialer.Dial(context.Background(), w.serverAddress+"/worker")
	if err != nil {
		return err
	}

	w.rpcClient = rpc2.NewClient(conn)

	go w.rpcClient.Run()

	w.rpcClient.Handle(proto.Common_Ping, w.Ping)

	var helloResp proto.HelloResp
	err = w.rpcClient.Call(proto.Control_Hello, proto.HelloReq{}, &helloResp)
	if err != nil {
		return err
	}

	for {
		time.Sleep(1 * time.Hour)
	}
}

var (
	_ proto.WorkerService = &Worker{}
)

func NewWorker(serverAddress string, serverCertificate string) *Worker {
	return &Worker{
		serverCertificate: serverCertificate,
		serverAddress:     serverAddress,
	}
}
