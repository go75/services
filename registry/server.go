package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

const (
	serviceAddr = "127.0.0.1:20000"
)

type RegistryService struct {
	serviceInfos            *serviceTable
	heartBeatWorkerNumber   int
	heartBeatAttempCount    int
	heartBeatAttempDuration time.Duration
	heartBeatCheckDuration  time.Duration
}

func Default() *RegistryService {
	return New(3, 3, time.Second, 30*time.Second)
}

func New(heartBeatWorkerNumber, heartBeatAttempCount int, heartBeatAttempDuration, heartBeatCheckDuration time.Duration) *RegistryService {
	return &RegistryService{
		serviceInfos:            newServiceTable(),
		heartBeatWorkerNumber:   heartBeatWorkerNumber,
		heartBeatAttempCount:    heartBeatAttempCount,
		heartBeatAttempDuration: heartBeatAttempDuration,
		heartBeatCheckDuration:  heartBeatCheckDuration,
	}
}

func (s *RegistryService) Run() error {
	go s.heartBeat()

	http.HandleFunc("/services", func(w http.ResponseWriter, r *http.Request) {
		statusCode := http.StatusOK
		switch r.Method {
		case http.MethodPost:
			serviceInfo, err := buildServiceInfo(r.Body)
			if err != nil {
				log.Println("build service info err:", err)
				statusCode = http.StatusInternalServerError
				goto END
			}

			err = s.regist(serviceInfo)
			if err != nil {
				log.Println("regist service err: ", err)
				statusCode = http.StatusInternalServerError
				goto END
			}

			serviceInfos := s.serviceInfos.buildRequiredServiceInfos(serviceInfo)
			data, err := json.Marshal(serviceInfos)
			if err != nil {
				log.Println("marshal srevice infos err: ", err)
				statusCode = http.StatusInternalServerError
				goto END
			}
			defer w.Write(data)


		case http.MethodDelete:
			serviceInfo, err := buildServiceInfo(r.Body)
			if err != nil {
				log.Println("build service info err:", err)
				statusCode = http.StatusInternalServerError
				goto END
			}

			s.unregist(serviceInfo)
			if err != nil {
				log.Println("unregist service err: ", err)
				statusCode = http.StatusInternalServerError
				goto END
			}

		default:
			statusCode = http.StatusMethodNotAllowed
			goto END
		}

	END:
		w.WriteHeader(statusCode)
	})

	return http.ListenAndServe(serviceAddr, nil)
}

func (s *RegistryService) heartBeat() {
	channel := make(chan *ServiceInfo, 1)
	for i := 0; i < s.heartBeatWorkerNumber; i++ {
		go func() {
			for service := range channel {
				for j := 0; j < s.heartBeatAttempCount; j++ {
					resp, err := http.Get("http://" + service.Addr + "/heart-beat")
					if err == nil && resp.StatusCode == http.StatusOK {
						goto NEXT
					}
					time.Sleep(s.heartBeatAttempDuration)
				}

				s.unregist(service)

			NEXT:
			}
		}()
	}

	for {
		s.serviceInfos.lock.RLock()
		for _, serviceInfos := range s.serviceInfos.serviceInfos {
			for i := len(serviceInfos) - 1; i >= 0; i-- {
				channel <- serviceInfos[i]
			}
		}
		s.serviceInfos.lock.RUnlock()
		time.Sleep(s.heartBeatCheckDuration)
	}
}

func (s *RegistryService) regist(service *ServiceInfo) error {
	s.serviceInfos.add(service)
	return s.notify(http.MethodPost, service)
}

func (s *RegistryService) unregist(service *ServiceInfo) error {
	s.serviceInfos.remove(service)
	return s.notify(http.MethodDelete, service)
}

func (s *RegistryService) notify(method string, serviceInfo *ServiceInfo) error {
	if method != http.MethodPost && method != http.MethodDelete {
		return fmt.Errorf("method not allowed with method: %s", method)
	}

	s.serviceInfos.lock.RLock()
	defer s.serviceInfos.lock.RUnlock()

	data, err := json.Marshal(serviceInfo)
	if err != nil {
		return err
	}

	for _, services := range s.serviceInfos.serviceInfos {
		for _, service := range services {
			for _, requiredServiceName := range service.RequiredServices {
				if requiredServiceName == serviceInfo.Name {
					req, err := http.NewRequest(method, "http://" + service.Addr + "/services", bytes.NewReader(data))
					if err != nil {
						log.Println("create http request with url http://" + service.Addr + "/services err:", err)
						continue
					}
					_, err = http.DefaultClient.Do(req)
					if err != nil {
						log.Println("nogify http://" + service.Addr + "/services err:", err)
						continue
					}
					log.Println("update url: ", service.Addr + "/services")
				}
			}
		}
	}

	return nil
}