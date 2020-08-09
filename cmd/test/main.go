package main

import (
	"flag"
	"fmt"
	"net"

	"github.com/arovesto/game-server/internal/config"
)

var cfgPath  = flag.String("config", "config.toml", "path to config")

func init() {
	flag.Parse()
}

func main() {
	cfg, err := config.LoadConfig(*cfgPath)
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}
 	listener, err := net.Listen("tcp", cfg.General.Address)
 	fmt.Println(listener)
}
