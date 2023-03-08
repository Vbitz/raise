package proto

import (
	"github.com/cenkalti/rpc2"
)

var (
	Common_Ping        = "Common_Ping"
	Control_Hello      = "Control_Hello"
	Common_SendMessage = "Common_SendMessage"
	Client_GetWorkers  = "Client_GetWorkers"
	Common_GetInfo     = "Common_GetInfo"
)

type MessageKind string

var (
	MessageReadFile  MessageKind = "Msg_ReadFile"
	MessageWriteFile MessageKind = "Msg_WriteFile"
	MessageRunScript MessageKind = "Msg_RunScript"
)

type PingReq struct {
	Name string
}

type PingResp struct {
	Message string
}

type SendMessageReq struct {
	Target string
	Kind   MessageKind

	Filename string
	Content  []byte
}

type SendMessageResp struct {
	Content []byte
}

type GetWorkersReq struct {
}

type GetWorkersResp struct {
	Workers []string
}

type GetInfoReq struct {
	Name string
}

type GetInfoResp struct {
	Hostname        string `json:"hostname"`
	HomeDir         string `json:"home"`
	OperatingSystem string `json:"os"`
	Architecture    string `json:"arch"`
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
	GetInfo(client *rpc2.Client, req GetInfoReq, resp *GetInfoResp) error
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
	GetInfo(client *rpc2.Client, req GetInfoReq, resp *GetInfoResp) error
}
