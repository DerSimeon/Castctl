package main

import (
	"github.com/simeon/castctl/cmd"
)

// version is stamped at build time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	cmd.Execute(version)
}
