package client

import (
	"fmt"

	"github.com/Vbitz/raise/v2/pkg/proto"
	"go.starlark.net/starlark"
)

type Remote struct {
	client *Client
	name   string
}

func (r *Remote) Ping(name string) (string, error) {
	var resp proto.PingResp

	err := r.client.rpcClient.Call(proto.Common_Ping, proto.PingReq{Name: name}, &resp)
	if err != nil {
		return "", err
	}

	return resp.Message, nil
}

func (r *Remote) Attr(name string) (starlark.Value, error) {
	if name == "ping" {
		return starlark.NewBuiltin("Remote.ping", func(
			thread *starlark.Thread,
			fn *starlark.Builtin,
			args starlark.Tuple,
			kwargs []starlark.Tuple,
		) (starlark.Value, error) {
			msg, err := r.Ping(r.name)
			if err != nil {
				return starlark.None, err
			}

			return starlark.String(msg), nil
		}), nil
	} else {
		return nil, nil
	}
}

func (*Remote) AttrNames() []string {
	return []string{"ping"}
}

func (*Remote) String() string       { return "Remote" }
func (*Remote) Truth() starlark.Bool { return starlark.True }
func (*Remote) Type() string         { return "Remote" }
func (*Remote) Freeze()              {}
func (*Remote) Hash() (uint32, error) {
	return 0, fmt.Errorf("Remote is unhashable")
}

var (
	_ starlark.Value    = &Remote{}
	_ starlark.HasAttrs = &Remote{}
)
