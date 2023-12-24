package main

import (
	"log"
	"services/registry"
	"services/service"
	"services/visistservice"
)

func main() {
	visistservice.Init()
	err := service.Run(&registry.ServiceInfo{
		Name:      "visist",
		Addr:      "127.0.0.1:20003",
		RequiredServices: []string{"log"},
	})
	if err != nil {
		log.Fatalln(err)
	}
}
