package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func registerMonitorHandler() {
	http.HandleFunc("/services", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		switch r.Method {
		case http.MethodPost:
			registration, err := buildRegistration(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			provider.add(registration)
			fmt.Printf("add service %s\n", registration.ServiceName)

		case http.MethodDelete:
			registration, err := buildRegistration(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			provider.remove(registration)
			fmt.Printf("remove service %s\n", registration.ServiceName)

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/heart-beat", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func RegistService(registration *Registration) error {
	registerMonitorHandler()

	data, err := json.Marshal(registration)
	if err != nil {
		return err
	}

	resp, err := http.Post("http://" + serviceAddr + "/services", "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Regist %s error with code %d", registration.ServiceName, resp.StatusCode)
	}

	err = provider.parseServiceInfos(resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func UnregistService(registration *Registration) error {
	data, err := json.Marshal(registration)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodDelete, "http://" + serviceAddr + "/services", bytes.NewReader(data))
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Unregist %s error with code %d", registration.ServiceName, resp.StatusCode)
	}

	return nil
}

var provider = newServiceTable()

func Get(serviceName string) string {
	reg := provider.get(serviceName)
	if reg != nil {
		return reg.ServiceAddr
	}
	return ""
}