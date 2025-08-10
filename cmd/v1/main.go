package main

import (
	"github.com/jfeliu007/goplantuml/pkg/client"
	"github.com/jfeliu007/goplantuml/pkg/config"
)

func main() {
	conf := config.NewBaseConfig()
	conf.LoadFromEnv()
	client.ClientForLegacyModeV1(conf)
}
