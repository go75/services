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
		ServiceName:      "visist",
		ServiceAddr:      "127.0.0.1:20004",
		RequiredServices: []string{"log"},
	})
	if err != nil {
		log.Fatalln(err)
	}
}
