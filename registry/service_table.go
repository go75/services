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
	serviceInfos map[string][]*Registration
	lock *sync.RWMutex
}

func newServiceTable() *serviceTable {
	return &serviceTable{
		serviceInfos: make(map[string][]*Registration),
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

func (t *serviceTable) buildRequiredServiceInfos(registration *Registration) map[string][]*Registration {
	m := make(map[string][]*Registration, len(registration.RequiredServices))
	t.lock.RLock()
	defer t.lock.RUnlock()
	
	for _, serviceName := range registration.RequiredServices {
		m[serviceName] = t.serviceInfos[serviceName]
	}

	return m
}

func (t *serviceTable) notify(method string, registration *Registration) error {
	if method != http.MethodPost && method != http.MethodDelete {
		fmt.Println(method, method == http.MethodPost, method == http.MethodDelete)
		return fmt.Errorf("Method not allowed with method: %s", method)
	}

	t.lock.RLock()
	defer t.lock.RUnlock()

	data, err := json.Marshal(registration)
	if err != nil {
		return err
	}

	for _, registrations := range t.serviceInfos {
		for _, reg := range registrations {
			for _, requiredServiceName := range reg.RequiredServices {
				if requiredServiceName == registration.ServiceName {
					req, err := http.NewRequest(method, "http://" + reg.ServiceAddr + "/services", bytes.NewReader(data))
					if err != nil {
						continue
					}
					log.Println("update url: ", reg.ServiceAddr + "/services")
					http.DefaultClient.Do(req)
				}
			}
		}
	}

	return nil
}

func (t *serviceTable) add(registration *Registration) {
	t.lock.Lock()
	defer t.lock.Unlock()

	log.Printf("Service table add %s with address %s\n", registration.ServiceName, registration.ServiceAddr)
	if registrations, ok := t.serviceInfos[registration.ServiceName]; ok {
		registrations = append(registrations, registration)
	} else {
		t.serviceInfos[registration.ServiceName] = []*Registration{registration}
	}
}

func (t *serviceTable) remove(registration *Registration) {
	t.lock.Lock()
	defer t.lock.Unlock()

	log.Printf("Service table remove %s with address %s\n", registration.ServiceName, registration.ServiceAddr)
	if registrations, ok := t.serviceInfos[registration.ServiceName]; ok {
		for i := len(registrations) - 1; i >= 0; i-- {
			if registrations[i].ServiceAddr == registration.ServiceAddr {
				registrations = append(registrations[:i], registrations[i+1:]...)
			}
		}
	}
}

func (t *serviceTable) get(serviceName string) *Registration {
	t.lock.RLock()
	defer t.lock.RUnlock()
	regs := t.serviceInfos[serviceName]
	return regs[rand.Intn(len(regs))]
}