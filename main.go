package main

import (
	"hmerritt/reactenv/command"
	"hmerritt/reactenv/version"
)

func main() {
	version.PrintTitle()
	command.Run()
}
