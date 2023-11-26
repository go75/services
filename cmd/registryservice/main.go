package main

import (
	"log"
	"services/registry"
)

func main() {
	registryService := registry.Default()
	err := registryService.Run()
	if err != nil {
		log.Fatalln(err)
	}
}
