package main

import (
	"generator_boilerplate/constant"
	"generator_boilerplate/server"
)

func main() {
	port := constant.Port
	ser := server.NewServer(port)
	ser.Start()
}
