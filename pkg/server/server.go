package server

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/Vbitz/raise/v2/pkg/proto"
	"github.com/cenkalti/rpc2"
	"github.com/gobwas/ws"
)

type Client struct {
	Name string
	// Client certificate as Base64 encoded DER format.
	CertificateString string

	server      *Server
	certificate *x509.Certificate
}

// Ping implements proto.ClientService
func (*Client) Ping(client *rpc2.Client, req proto.PingReq, resp *proto.PingResp) error {
	if req.Name != "" {
		return fmt.Errorf("not implemented")
	}

	*resp = proto.PingResp{}

	return nil
}

var (
	_ proto.ClientService = &Client{}
)

type Worker struct {
	server    *Server
	addr      string
	rpcServer *rpc2.Server
	rpcClient *rpc2.Client
}

// Hello implements proto.ControlService
func (w *Worker) Hello(client *rpc2.Client, req proto.HelloReq, resp *proto.HelloResp) error {
	var pingResp proto.PingResp

	err := client.Call(proto.Common_Ping, proto.PingReq{}, &pingResp)
	if err != nil {
		return err
	}

	log.Printf("worker from %s registered", w.addr)

	w.rpcClient = client

	return nil
}

// Ping implements proto.ControlService
func (*Worker) Ping(client *rpc2.Client, req proto.PingReq, resp *proto.PingResp) error {
	if req.Name != "" {
		return fmt.Errorf("not implemented")
	}

	*resp = proto.PingResp{}

	return nil
}

var (
	_ proto.ClientService  = &Worker{}
	_ proto.ControlService = &Worker{}
)

type Server struct {
	addr             string
	certFile         string
	keyFile          string
	permittedClients []*Client
	mux              *http.ServeMux
	upgrader         ws.HTTPUpgrader
	connectedWorkers []*Worker
}

func (s *Server) authenticateClient(certs []*x509.Certificate) *Client {
	// log.Printf("certs = %v", certs)
	for _, cert := range certs {
		for _, client := range s.permittedClients {
			// log.Printf("cert = %v client = %v", cert, client)
			// TODO(joshua): Is this the right way to validate client certificates.
			if cert.Equal(client.certificate) {
				return client
			}
		}
	}
	return nil
}

func (s *Server) Listen() error {
	inner, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	cert, err := tls.LoadX509KeyPair(s.certFile, s.keyFile)
	if err != nil {
		return err
	}

	config := &tls.Config{
		ClientAuth:   tls.RequestClientCert,
		Certificates: []tls.Certificate{cert},
	}
	tlsListener := tls.NewListener(inner, config)

	log.Printf("Listening on %s", inner.Addr().String())

	return http.Serve(tlsListener, s.mux)
}

func (s *Server) AddClient(client Client) error {
	bytes, err := base64.StdEncoding.DecodeString(client.CertificateString)
	if err != nil {
		return err
	}
	cert, err := x509.ParseCertificate(bytes)
	if err != nil {
		return err
	}

	s.permittedClients = append(s.permittedClients, &Client{
		Name:              client.Name,
		CertificateString: client.CertificateString,
		certificate:       cert,
	})

	return nil
}

func (s *Server) handleClient(w http.ResponseWriter, r *http.Request) {
	client := s.authenticateClient(r.TLS.PeerCertificates)

	if client == nil {
		log.Printf("client failed authentication from: %s", r.RemoteAddr)

		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "Unauthorised")
		return
	}

	log.Printf("client connected from: %s", r.RemoteAddr)

	conn, _, _, err := s.upgrader.Upgrade(r, w)
	if err != nil {
		log.Printf("error creating codec: %v", err)
	}
	defer conn.Close()

	client.server = s

	server := rpc2.NewServer()

	server.Handle(proto.Common_Ping, client.Ping)

	server.ServeConn(conn)
}

func (s *Server) handleWorker(w http.ResponseWriter, r *http.Request) {
	log.Printf("worker connected from: %s", r.RemoteAddr)

	conn, _, _, err := s.upgrader.Upgrade(r, w)
	if err != nil {
		log.Printf("error creating codec: %v", err)
	}
	defer conn.Close()

	worker := &Worker{
		server:    s,
		addr:      r.RemoteAddr,
		rpcServer: rpc2.NewServer(),
	}

	s.connectedWorkers = append(s.connectedWorkers, worker)

	worker.rpcServer.Handle(proto.Common_Ping, worker.Ping)
	worker.rpcServer.Handle(proto.Control_Hello, worker.Hello)

	worker.rpcServer.ServeConn(conn)
}

func NewServer(addr string, certFile string, keyFile string) *Server {
	s := &Server{
		addr:     addr,
		certFile: certFile,
		keyFile:  keyFile,
		upgrader: ws.HTTPUpgrader{},
		mux:      http.NewServeMux(),
	}

	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK")
	})
	s.mux.HandleFunc("/client", s.handleClient)
	s.mux.HandleFunc("/worker", s.handleWorker)

	return s
}
