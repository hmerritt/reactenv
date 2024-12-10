package main

import (
	"github.com/hmerritt/reactenv/command"
	"github.com/hmerritt/reactenv/version"
)

func main() {
	version.PrintTitle()
	command.Run()
}
