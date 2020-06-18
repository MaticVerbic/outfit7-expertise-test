package main

import (
	"expertisetest/config"
	"expertisetest/server"
)

func main() {
	config.GetInstance()
	server.New().Serve()
}
