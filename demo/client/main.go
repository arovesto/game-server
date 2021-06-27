package main

import (
	"github.com/arovesto/gio/client"
	_ "github.com/arovesto/gio/demo/entities"
)

func main() {
	client.RunClient(60, "/static/assets")
}
