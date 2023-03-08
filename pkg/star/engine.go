package star

import (
	"go.starlark.net/starlark"

	"github.com/Vbitz/raise/v2/pkg/client"
	"github.com/Vbitz/raise/v2/pkg/star/builtin"
)

type StarEngine struct {
}

func (e *StarEngine) newThread(name string) *starlark.Thread {
	return &starlark.Thread{Name: name}
}

func (e *StarEngine) RunFile(client *client.Client, remote *client.Remote, filename string, fileContents []byte) error {
	thread := e.newThread(filename)

	builtin := builtin.Globals

	builtin["client"] = client
	if remote != nil {
		builtin["remote"] = remote
	}

	_, err := starlark.ExecFile(thread, filename, fileContents, builtin)
	if err != nil {
		return err
	}

	return nil
}

func NewEngine() *StarEngine {
	return &StarEngine{}
}
