package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"sync"
)

type serviceTable struct {
	serviceInfos map[string][]*ServiceInfo
	lock *sync.RWMutex
}

func newServiceTable() *serviceTable {
	return &serviceTable{
		serviceInfos: make(map[string][]*ServiceInfo),
		lock: new(sync.RWMutex),
	}
}

func (t *serviceTable) parseServiceInfos(reader io.ReadCloser) (err error){
	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	defer func() {
		err = reader.Close()
	}()
	t.lock.Lock()
	defer t.lock.Unlock()
	err = json.Unmarshal(data, &t.serviceInfos)
	return
}

func (t *serviceTable) buildRequiredServiceInfos(service *ServiceInfo) map[string][]*ServiceInfo {
	m := make(map[string][]*ServiceInfo, len(service.RequiredServices))
	t.lock.RLock()
	defer t.lock.RUnlock()
	
	for _, serviceName := range service.RequiredServices {
		m[serviceName] = t.serviceInfos[serviceName]
	}

	return m
}

func (t *serviceTable) notify(method string, service *ServiceInfo) error {
	if method != http.MethodPost && method != http.MethodDelete {
		fmt.Println(method, method == http.MethodPost, method == http.MethodDelete)
		return fmt.Errorf("Method not allowed with method: %s", method)
	}

	t.lock.RLock()
	defer t.lock.RUnlock()

	data, err := json.Marshal(service)
	if err != nil {
		return err
	}

	for _, services := range t.serviceInfos {
		for _, service := range services {
			for _, requiredServiceName := range service.RequiredServices {
				if requiredServiceName == service.Name {
					req, err := http.NewRequest(method, "http://" + service.Addr + "/services", bytes.NewReader(data))
					if err != nil {
						continue
					}
					log.Println("update url: ", service.Addr + "/services")
					http.DefaultClient.Do(req)
				}
			}
		}
	}

	return nil
}

func (t *serviceTable) add(service *ServiceInfo) {
	t.lock.Lock()
	defer t.lock.Unlock()

	log.Printf("Service table add %s with address %s\n", service.Name, service.Addr)
	if services, ok := t.serviceInfos[service.Name]; ok {
		services = append(services, service)
	} else {
		t.serviceInfos[service.Name] = []*ServiceInfo{service}
	}
}

func (t *serviceTable) remove(service *ServiceInfo) {
	t.lock.Lock()
	defer t.lock.Unlock()

	log.Printf("Service table remove %s with address %s\n", service.Name, service.Addr)
	if services, ok := t.serviceInfos[service.Name]; ok {
		for i := len(services) - 1; i >= 0; i-- {
			if services[i].Addr == service.Addr {
				services = append(services[:i], services[i+1:]...)
			}
		}
	}
}

func (t *serviceTable) get(serviceName string) *ServiceInfo {
	t.lock.RLock()
	defer t.lock.RUnlock()
	services, ok := t.serviceInfos[serviceName]
	if !ok && len(services) < 1 {
		return nil
	}
	idx := rand.Intn(len(services))
	return services[idx]
}