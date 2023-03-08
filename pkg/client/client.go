package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"net"
	"os"

	"github.com/Vbitz/raise/v2/pkg/proto"
	"github.com/cenkalti/rpc2"
	"github.com/gobwas/ws"
	"go.starlark.net/starlark"
)

type Client struct {
	serverAddress     string
	serverCertificate string
	rpcClient         *rpc2.Client
	rpcConn           net.Conn
	clientCertificate string
	clientKey         string
}

// Ping implements proto.CommonService
func (c *Client) Ping(client *rpc2.Client, req proto.PingReq, resp *proto.PingResp) error {
	*resp = proto.PingResp{}
	return nil
}

func (c *Client) Connect() error {
	dialer := ws.Dialer{}

	certPool := x509.NewCertPool()

	pemBytes, err := base64.StdEncoding.DecodeString(c.serverCertificate)
	if err != nil {
		return fmt.Errorf("failed to decode server certificate: %v", err)
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
		return fmt.Errorf("failed to dial server: %v", err)
	}

	c.rpcConn = conn

	c.rpcClient = rpc2.NewClient(c.rpcConn)

	go c.rpcClient.Run()

	c.rpcClient.Handle(proto.Common_Ping, c.Ping)

	return nil
}

func (c *Client) Close() error {
	if c.rpcConn != nil {
		if err := c.rpcConn.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) Remote(name string) (*Remote, error) {
	// Lazily connect to the server when we get our first client connection.
	if c.rpcConn == nil {
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

func (c *Client) GetWorkers() ([]string, error) {
	// Lazily connect to the server when we get our first client connection.
	if c.rpcConn == nil {
		err := c.Connect()
		if err != nil {
			return nil, err
		}
	}

	var resp proto.GetWorkersResp
	err := c.rpcClient.Call(proto.Client_GetWorkers, proto.GetWorkersReq{}, &resp)
	if err != nil {
		return nil, err
	}

	return resp.Workers, nil
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
	} else if name == "get_workers" {
		return starlark.NewBuiltin("Client.get_workers", func(
			thread *starlark.Thread,
			fn *starlark.Builtin,
			args starlark.Tuple,
			kwargs []starlark.Tuple,
		) (starlark.Value, error) {
			workerList, err := c.GetWorkers()
			if err != nil {
				return starlark.None, err
			}

			var ret []starlark.Value

			for _, name := range workerList {
				ret = append(ret, starlark.String(name))
			}

			return starlark.NewList(ret), nil
		}), nil
		// } else if name == "read_file" {
		// } else if name == "write_file" {
	} else {
		return nil, nil
	}
}

func (*Client) AttrNames() []string {
	return []string{"remote", "get_workers", "read_file", "write_file"}
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

	_ proto.CommonService = &Client{}
)

func NewClient(serverAddress string, serverCertificate string, clientCertificate string, clientKey string) *Client {
	return &Client{
		serverAddress:     serverAddress,
		serverCertificate: serverCertificate,
		clientCertificate: clientCertificate,
		clientKey:         clientKey,
	}
}
