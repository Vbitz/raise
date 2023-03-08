package client

import (
	"encoding/json"
	"fmt"

	"github.com/Vbitz/raise/v2/pkg/proto"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkjson"
)

type Remote struct {
	client *Client
	name   string
	info   *proto.GetInfoResp
}

func (r *Remote) Ping() (string, error) {
	var resp proto.PingResp

	err := r.client.rpcClient.Call(proto.Common_Ping, proto.PingReq{Name: r.name}, &resp)
	if err != nil {
		return "", err
	}

	return resp.Message, nil
}

func (r *Remote) GetInfo() (*proto.GetInfoResp, error) {
	if r.info != nil {
		return r.info, nil
	}

	// Make the RPC call to the server to get the worker info.
	err := r.client.rpcClient.Call(proto.Common_GetInfo, proto.GetInfoReq{Name: r.name}, &r.info)
	if err != nil {
		return nil, fmt.Errorf("failed to call GetInfo: %v", err)
	}

	return r.info, nil
}

func (r *Remote) ReadFile(filename string) ([]byte, error) {
	var resp proto.SendMessageResp

	err := r.client.rpcClient.Call(proto.Common_SendMessage, proto.SendMessageReq{
		Target:   r.name,
		Kind:     proto.MessageReadFile,
		Filename: filename,
	}, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to call ReadFile: %v", err)
	}

	return resp.Content, nil
}

func (r *Remote) WriteFile(filename string, content []byte) error {
	var resp proto.SendMessageResp

	err := r.client.rpcClient.Call(proto.Common_SendMessage, proto.SendMessageReq{
		Target:   r.name,
		Kind:     proto.MessageWriteFile,
		Filename: filename,
		Content:  content,
	}, &resp)
	if err != nil {
		return fmt.Errorf("failed to call WriteFile: %v", err)
	}

	return nil
}

func (r *Remote) RunScript(script string) ([]byte, error) {
	var resp proto.SendMessageResp

	err := r.client.rpcClient.Call(proto.Common_SendMessage, proto.SendMessageReq{
		Target:  r.name,
		Kind:    proto.MessageRunScript,
		Content: []byte(script),
	}, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to call WriteFile: %v", err)
	}

	return resp.Content, nil
}

func (r *Remote) Attr(name string) (starlark.Value, error) {
	if name == "ping" {
		return starlark.NewBuiltin("Remote.ping", func(
			thread *starlark.Thread,
			fn *starlark.Builtin,
			args starlark.Tuple,
			kwargs []starlark.Tuple,
		) (starlark.Value, error) {
			msg, err := r.Ping()
			if err != nil {
				return starlark.None, err
			}

			return starlark.String(msg), nil
		}), nil
	} else if name == "info" {
		return starlark.NewBuiltin("Remote.info", func(
			thread *starlark.Thread,
			fn *starlark.Builtin,
			args starlark.Tuple,
			kwargs []starlark.Tuple,
		) (starlark.Value, error) {
			info, err := r.GetInfo()
			if err != nil {
				return starlark.None, err
			}

			infoJson, err := json.Marshal(info)
			if err != nil {
				return starlark.None, err
			}

			// Get the json.decode method from Starlark.
			// This is a easy way of converting an arbitrary JSON structure info a Starlark value.
			loadFunc := starlarkjson.Module.Members["decode"].(*starlark.Builtin)

			ret, err := loadFunc.CallInternal(thread, []starlark.Value{starlark.String(infoJson)}, []starlark.Tuple{})
			if err != nil {
				return starlark.None, err
			}

			return ret, nil
		}), nil
	} else if name == "read_file" {
		return starlark.NewBuiltin("Remote.read_file", func(
			thread *starlark.Thread,
			fn *starlark.Builtin,
			args starlark.Tuple,
			kwargs []starlark.Tuple,
		) (starlark.Value, error) {
			var (
				filename string
			)
			if err := starlark.UnpackArgs("Remote.read_file", args, kwargs,
				"filename", &filename,
			); err != nil {
				return starlark.None, err
			}

			content, err := r.ReadFile(filename)
			if err != nil {
				return starlark.None, err
			}

			return starlark.String(content), nil
		}), nil
	} else if name == "write_file" {
		return starlark.NewBuiltin("Remote.write_file", func(
			thread *starlark.Thread,
			fn *starlark.Builtin,
			args starlark.Tuple,
			kwargs []starlark.Tuple,
		) (starlark.Value, error) {
			var (
				filename string
				content  string
			)
			if err := starlark.UnpackArgs("Remote.write_file", args, kwargs,
				"filename", &filename,
				"content", &content,
			); err != nil {
				return starlark.None, err
			}

			err := r.WriteFile(filename, []byte(content))
			if err != nil {
				return starlark.None, err
			}

			return starlark.None, nil
		}), nil
	} else if name == "run_script" {
		return starlark.NewBuiltin("Remote.run_script", func(
			thread *starlark.Thread,
			fn *starlark.Builtin,
			args starlark.Tuple,
			kwargs []starlark.Tuple,
		) (starlark.Value, error) {
			var (
				script string
			)
			if err := starlark.UnpackArgs("Remote.run_script", args, kwargs,
				"script", &script,
			); err != nil {
				return starlark.None, err
			}

			result, err := r.RunScript(script)
			if err != nil {
				return starlark.None, err
			}

			return starlark.String(result), nil
		}), nil
	} else {
		return nil, nil
	}
}

func (*Remote) AttrNames() []string {
	return []string{"ping", "info", "read_file", "write_file", "run_script"}
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
