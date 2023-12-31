package main

import (
	"log"
	"services/logservice"
	"services/registry"
	"services/service"
)

func main() {
	logservice.Init("./services.log")
	err := service.Run(&registry.ServiceInfo{
		Name:      "log",
		Addr:      "127.0.0.1:20002",
		RequiredServices: make([]string, 0),
	})
	if err != nil {
		log.Fatalln(err)
	}
}
