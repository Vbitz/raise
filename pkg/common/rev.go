// From: https://icinga.com/blog/2022/05/25/embedding-git-commit-information-in-go-binaries/
package common

import (
	_ "embed"
)

//go:generate sh -c "printf %s $(git rev-parse HEAD) > commit.txt"
//go:embed commit.txt
var Commit string
