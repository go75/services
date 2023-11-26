package logservice

import (
	"bytes"
	"fmt"
	"net/http"
	"services/registry"
)

func Println(registration *registry.Registration, s string) error {
	resp, err := http.Post("http://"+registration.ServiceAddr+"/log", "text/plain", bytes.NewReader([]byte(s)))
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Response Error with code: %d", resp.StatusCode)
	}
	return nil
}
