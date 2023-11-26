package main

import (
	"log"
	"services/registry"
	"services/service"
	"services/visistservice"
)

func main() {
	visistservice.Init()
	err := service.Run(&registry.Registration{
		ServiceName:      "VisistService",
		ServiceAddr:      "127.0.0.1:20003",
		RequiredServices: []string{"LogService"},
	})
	if err != nil {
		log.Fatalln(err)
	}
}
