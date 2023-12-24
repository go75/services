package registry

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
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

func (t *serviceTable) add(service *ServiceInfo) {
	t.lock.Lock()
	defer t.lock.Unlock()

	log.Printf("Service table add %s with address %s\n", service.Name, service.Addr)
	t.serviceInfos[service.Name] = append(t.serviceInfos[service.Name], service)
}

func (t *serviceTable) remove(service *ServiceInfo) {
	t.lock.Lock()
	defer t.lock.Unlock()

	log.Printf("Service table remove %s with address %s\n", service.Name, service.Addr)
	services := t.serviceInfos[service.Name]
	for i := len(services) - 1; i >= 0; i-- {
		if services[i].Addr == service.Addr {
			t.serviceInfos[service.Name] = append(services[:i], services[i+1:]...)
		}
	}
}

func (t *serviceTable) get(serviceName string) *ServiceInfo {
	t.lock.RLock()
	defer t.lock.RUnlock()
	services, ok := t.serviceInfos[serviceName]
	if !ok || len(services) < 1 {
		return nil
	}
	idx := rand.Intn(len(services))
	return services[idx]
}

func (t *serviceTable) dump() {
	t.lock.RLock()
	defer t.lock.RUnlock()
	fmt.Println("==========Dump Service Table Start==========")
	for k, v := range t.serviceInfos {
		fmt.Print("Service " + k + ": [ ")
		for i := 0; i < len(v); i++ {
			fmt.Print(v[i].Addr + " ")
		}
		fmt.Println("]")
	}
	fmt.Println("==========Dump Service Table End==========")
}