package proto

import (
	"github.com/cenkalti/rpc2"
)

var (
	Common_Ping        = "Common_Ping"
	Control_Hello      = "Control_Hello"
	Common_SendMessage = "Common_SendMessage"
)

type PingReq struct {
	Name string
}

type PingResp struct {
	Message string
}

type SendMessageReq struct {
	Target string
}

type SendMessageResp struct {
}

type HelloReq struct {
	Name string
}

type HelloResp struct{}

type CommonService interface {
	Ping(client *rpc2.Client, req PingReq, resp *PingResp) error
}

// Client -> Server Communication
type ClientService interface {
	CommonService

	SendMessage(client *rpc2.Client, req SendMessageReq, resp *SendMessageResp) error
}

// Worker -> Server Communication
type ControlService interface {
	CommonService

	Hello(client *rpc2.Client, req HelloReq, resp *HelloResp) error
}

// Server -> Worker Communication
type WorkerService interface {
	CommonService

	SendMessage(client *rpc2.Client, req SendMessageReq, resp *SendMessageResp) error
}
