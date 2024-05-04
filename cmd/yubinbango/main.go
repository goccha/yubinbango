package main

import (
	"github.com/goccha/yubinbango/internal/cmd"
	"github.com/goccha/yubinbango/pkg/env"
)

var (
	version  = "v0.0.0"
	revision = "0000000"
)

// main
func main() {
	env.Setup(version, revision)
	cmd.Execute()
}
