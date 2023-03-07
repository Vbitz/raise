package proto

import (
	"github.com/cenkalti/rpc2"
)

var (
	Common_Ping   = "Common_Ping"
	Control_Hello = "Control_Hello"
)

type PingReq struct {
	Name string
}

type PingResp struct{}

type HelloReq struct {
}

type HelloResp struct{}

type CommonService interface {
	Ping(client *rpc2.Client, req PingReq, resp *PingResp) error
}

// Client -> Server Communication
type ClientService interface {
	CommonService
}

// Worker -> Server Communication
type ControlService interface {
	CommonService

	Hello(client *rpc2.Client, req HelloReq, resp *HelloResp) error
}

// Server -> Worker Communication
type WorkerService interface {
	CommonService
}
