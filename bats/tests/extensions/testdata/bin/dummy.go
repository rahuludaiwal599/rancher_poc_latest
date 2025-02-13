package main

import (
	"context"
	"log"
	"os"
	"os/exec"
)

func main() {
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, os.Args[1], os.Args[2:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err, ok := err.(*exec.ExitError); ok {
		if err.ExitCode() > -1 {
			os.Exit(err.ExitCode())
		}
	} else if err != nil {
		log.Fatal(err)
	}
}
