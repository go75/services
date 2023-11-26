package service

import (
	"context"
	"fmt"
	"net/http"
	"services/registry"
)

type Service interface {
	Init()
}

func Run(registration *registry.Registration) (err error) {
	err = registry.RegistService(registration)
	if err != nil {
		return err
	}
	defer func() {
		err = registry.UnregistService(registration)
	}()

	srv := http.Server{Addr: registration.ServiceAddr}

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
