package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"os"

	"capnproto.org/go/capnp/v3/rpc"
	"github.com/Vbitz/raise/v2/pkg/proto"
	"github.com/gobwas/ws"
	"go.starlark.net/starlark"
	wsproto "zenhack.net/go/websocket-capnp"
)

type Client struct {
	serverAddress     string
	serverCertificate string
	rpcConn           *rpc.Conn
	rpcClient         *proto.ClientService
	clientCertificate string
	clientKey         string
}

func (c *Client) Connect() error {
	dialer := ws.Dialer{}

	certPool := x509.NewCertPool()

	pemBytes, err := base64.StdEncoding.DecodeString(c.serverCertificate)
	if err != nil {
		return err
	}

	ok := certPool.AppendCertsFromPEM(pemBytes)
	if !ok {
		return fmt.Errorf("failed to add server certificate")
	}

	dialer.TLSConfig = &tls.Config{}

	certContent, err := os.ReadFile(c.clientCertificate)
	if err != nil {
		return fmt.Errorf("failed to read certificate: %v", err)
	}

	keyContent, err := os.ReadFile(c.clientKey)
	if err != nil {
		return fmt.Errorf("failed to read key: %v", err)
	}

	priv, err := x509.ParsePKCS1PrivateKey(keyContent)
	if err != nil {
		return fmt.Errorf("failed to parse key: %v", err)
	}

	crt := tls.Certificate{
		PrivateKey:  priv,
		Certificate: [][]byte{certContent},
	}

	dialer.TLSConfig.GetClientCertificate = func(
		cri *tls.CertificateRequestInfo,
	) (*tls.Certificate, error) {
		err := cri.SupportsCertificate(&crt)
		if err != nil {
			return nil, err
		}

		return &crt, nil
	}

	dialer.TLSConfig.RootCAs = certPool

	conn, _, _, err := dialer.Dial(context.Background(), c.serverAddress+"/client")
	if err != nil {
		return err
	}

	transport := wsproto.NewTransport(conn, false)
	c.rpcConn = rpc.NewConn(transport, nil)

	client := proto.ClientService(c.rpcConn.Bootstrap(context.Background()))

	c.rpcClient = &client

	return nil
}

func (c *Client) Close() error {
	if c.rpcClient != nil {
		c.rpcClient.Release()
	}
	if c.rpcConn != nil {
		if err := c.rpcConn.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) Remote(name string) (*Remote, error) {
	// Lazily connect to the server when we get our first client connection.
	if c.rpcClient == nil {
		err := c.Connect()
		if err != nil {
			return nil, err
		}
	}

	return &Remote{
		client: c,
		name:   name,
	}, nil
}

// Attr implements starlark.HasAttrs
func (c *Client) Attr(name string) (starlark.Value, error) {
	if name == "remote" {
		return starlark.NewBuiltin("Client.remote", func(
			thread *starlark.Thread,
			fn *starlark.Builtin,
			args starlark.Tuple,
			kwargs []starlark.Tuple,
		) (starlark.Value, error) {
			var (
				name string
			)
			if err := starlark.UnpackArgs("Client.remote", args, kwargs,
				"remote?", &name); err != nil {
				return starlark.None, err
			}

			return c.Remote(name)
		}), nil
		// } else if name == "read_file" {
		// } else if name == "write_file" {
	} else {
		return nil, nil
	}
}

func (*Client) AttrNames() []string {
	return []string{"remote", "read_file", "write_file"}
}

func (*Client) String() string       { return "Client" }
func (*Client) Truth() starlark.Bool { return starlark.True }
func (*Client) Type() string         { return "Client" }
func (*Client) Freeze()              {}
func (*Client) Hash() (uint32, error) {
	return 0, fmt.Errorf("Client is unhashable")
}

var (
	_ starlark.Value    = &Client{}
	_ starlark.HasAttrs = &Client{}
)

func NewClient(serverAddress string, serverCertificate string, clientCertificate string, clientKey string) *Client {
	return &Client{
		serverAddress:     serverAddress,
		serverCertificate: serverCertificate,
		clientCertificate: clientCertificate,
		clientKey:         clientKey,
	}
}
