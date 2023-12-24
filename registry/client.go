package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func registerMonitorHandler() {
	http.HandleFunc("/services", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			service, err := buildServiceInfo(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			provider.add(service)
			fmt.Printf("add service %s\n", service.Name)

		case http.MethodDelete:
			service, err := buildServiceInfo(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			provider.remove(service)
			fmt.Printf("remove service %s\n", service.Name)

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

func RegistService(service *ServiceInfo) error {
	registerMonitorHandler()

	data, err := json.Marshal(service)
	if err != nil {
		return err
	}

	resp, err := http.Post("http://"+serviceAddr+"/services", "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("regist %s error with code %d", service.Name, resp.StatusCode)
	}

	err = provider.parseServiceInfos(resp.Body)
	if err != nil {
		return err
	}

	provider.dump()

	return nil
}

func UnregistService(service *ServiceInfo) error {
	data, err := json.Marshal(service)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodDelete, "http://"+serviceAddr+"/services", bytes.NewReader(data))
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unregist %s error with code %d", service.Name, resp.StatusCode)
	}

	return nil
}

var provider = newServiceTable()

func Get(serviceName string) string {
	service := provider.get(serviceName)
	if service != nil {
		return service.Addr
	}
	return ""
}
