package main

import (
	"log"
	"services/gatewayservice"
	"services/registry"
	"services/service"
)

func main() {
	gatewayservice.Init()
	err := service.Run(&registry.ServiceInfo{
		Name: "gateway",
		Addr: "127.0.0.1:20001",
		RequiredServices: []string{"log", "visist"},
	})

	if err != nil {
		log.Fatalln(err)
	}
}