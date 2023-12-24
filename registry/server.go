package registry

import (
	"encoding/json"
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
				statusCode = http.StatusInternalServerError
				goto END
			}

			err = s.regist(serviceInfo)
			if err != nil {
				statusCode = http.StatusInternalServerError
				goto END
			}

			serviceInfos := s.serviceInfos.buildRequiredServiceInfos(serviceInfo)
			data, err := json.Marshal(&serviceInfos)
			if err != nil {
				statusCode = http.StatusInternalServerError
				goto END
			}
			defer w.Write(data)

		case http.MethodDelete:
			serviceInfo, err := buildServiceInfo(r.Body)
			if err != nil {
				statusCode = http.StatusInternalServerError
				goto END
			}

			s.unregist(serviceInfo)
			if err != nil {
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
	println("add " + service.Name + " after ->", s.serviceInfos.get(service.Name).Addr)
	return s.serviceInfos.notify(http.MethodPost, service)
}

func (s *RegistryService) unregist(service *ServiceInfo) error {
	s.serviceInfos.remove(service)
	println("remove " + service.Name + " after ->", len(s.serviceInfos.serviceInfos[service.Name]))
	return s.serviceInfos.notify(http.MethodDelete, service)
}
