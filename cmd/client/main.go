package main

import (
	"os"

	"github.com/jfeliu007/goplantuml/pkg/client"
)

func main() {
	cmd := client.InitRootCmdForGoPlantUMLV2()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
