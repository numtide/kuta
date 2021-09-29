package main

import (
	"log"
	"os"

	"github.com/numtide/kuta/mod"
)

func main() {
	var err error

	// Patch the system
	err = mod.PatchUser()
	if err != nil {
		log.Println("[kuta] patch-user: %w", err)
		os.Exit(1)
	}

	// Drop the kuma argument
	args := os.Args[1:]
	// Run the command
	status, err := mod.BashExec(args)
	if err != nil {
		log.Println("[kuta] bash-exec: %w", err)
		os.Exit(1)
	}

	os.Exit(status)
}
