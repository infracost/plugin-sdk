package main

import (
	"github.com/infracost/parser/internal/plugin"
	"github.com/infracost/parser/plugin/template/server"
)

func main() {
	plugin.Expose(server.New())
}
