package proto

import (
	"github.com/cenkalti/rpc2"
)

var (
	Common_Ping        = "Common_Ping"
	Control_Hello      = "Control_Hello"
	Common_SendMessage = "Common_SendMessage"
	Client_GetWorkers  = "Client_GetWorkers"
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

type GetWorkersReq struct {
}

type GetWorkersResp struct {
	Workers []string
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
	GetWorkers(client *rpc2.Client, req GetWorkersReq, resp *GetWorkersResp) error
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
