package server

import (
	context "context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"net/http"

	"capnproto.org/go/capnp/v3"
	"capnproto.org/go/capnp/v3/rpc"
	"capnproto.org/go/capnp/v3/rpc/transport"
	"github.com/Vbitz/raise/v2/pkg/proto"
	"github.com/gobwas/ws"
	wsproto "zenhack.net/go/websocket-capnp"
)

type Client struct {
	Name string
	// Client certificate as Base64 encoded DER format.
	CertificateString string

	certificate *x509.Certificate
}

type Server struct {
	addr             string
	certFile         string
	keyFile          string
	permittedClients []*Client
	mux              *http.ServeMux
	upgrader         ws.HTTPUpgrader
}

// Ping implements proto.ClientService_Server
func (*Server) Ping(ctx context.Context, m proto.Service_ping) error {
	name, err := m.Args().Name()
	if err != nil {
		return err
	}

	if name != "" {
		return fmt.Errorf("not implemented")
	}

	_, err = m.AllocResults()
	if err != nil {
		return err
	}

	return nil
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

	codec, err := wsproto.UpgradeHTTP(s.upgrader, r, w)
	if err != nil {
		log.Printf("error creating codec: %v", err)
	}
	defer codec.Close()

	protoClient := proto.ClientService_ServerToClient(s)

	conn := rpc.NewConn(transport.New(codec), &rpc.Options{
		// The BootstrapClient is the RPC interface that will be made available
		// to the remote endpoint by default.
		BootstrapClient: capnp.Client(protoClient),
	})
	defer conn.Close()

	select {
	case <-conn.Done():
		return
	}
}

func (s *Server) handleWorker(w http.ResponseWriter, r *http.Request) {
	codec, err := wsproto.UpgradeHTTP(s.upgrader, r, w)
	if err != nil {
		log.Printf("error creating codec: %v", err)
	}
	defer codec.Close()

	protoClient := proto.WorkerService_ServerToClient(s)

	conn := rpc.NewConn(transport.New(codec), &rpc.Options{
		// The BootstrapClient is the RPC interface that will be made available
		// to the remote endpoint by default.
		BootstrapClient: capnp.Client(protoClient),
	})
	defer conn.Close()

	select {
	case <-conn.Done():
		return
	}
}

var (
	_ proto.ClientService_Server = &Server{}
	_ proto.WorkerService_Server = &Server{}
)

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
