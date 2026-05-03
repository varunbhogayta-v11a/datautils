package main

import (
	"github.com/varunbhogayta-v11a/datautils/cmd"
	"os"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
