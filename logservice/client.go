package logservice

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"services/registry"
)

func Println(s string) error {
	serviceAddr := registry.Get("log")
	fmt.Println("service addr: " + serviceAddr)
	if serviceAddr == "" {
		return errors.New("No services are available")	
	}
	resp, err := http.Post("http://"+serviceAddr+"/log", "text/plain", bytes.NewReader([]byte(s)))
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Response Error with code: %d", resp.StatusCode)
	}
	return nil
}
