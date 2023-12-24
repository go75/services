package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"services/registry"
)

type Service interface {
	Init()
}

func Run(service *registry.ServiceInfo) (err error) {
	err = registry.RegistService(service)
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, registry.UnregistService(service))
	}()

	srv := http.Server{Addr: service.Addr}

	go func() {
		fmt.Println("Press any key to stop.")
		var s string
		fmt.Scan(&s)
		srv.Shutdown(context.Background())
	}()

	err = srv.ListenAndServe()
	if err != nil {
		return err
	}

	return nil
}
