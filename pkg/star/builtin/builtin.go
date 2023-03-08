package builtin

import (
	"fmt"
	"path"

	"go.starlark.net/starlark"
)

var Globals = starlark.StringDict{}

func init() {
	Globals["join"] = starlark.NewBuiltin("join", builtinJoin)
}

func builtinJoin(
	thread *starlark.Thread,
	builtin *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	var elements []string

	for _, arg := range args {
		str, ok := arg.(starlark.String)
		if !ok {
			return starlark.None, fmt.Errorf("expected %s, got %s", "String", arg.Type())
		}
		elements = append(elements, str.GoString())
	}

	return starlark.String(path.Join(elements...)), nil
}
