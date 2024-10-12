package main

import (
	"generator_boilerplate/constant"
	"generator_boilerplate/server"
)

func main() {
	url := constant.Url
	ser := server.NewServer(url)
	ser.Start()
}
