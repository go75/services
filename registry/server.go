package registry

import (
	"encoding/json"
	"net/http"
	"time"
)

const (
	serviceName = "Registry Service"
	serviceAddr = "127.0.0.1:20000"
)

type RegistryService struct {
	serviceInfos *serviceTable
	heartBeatWorkerNumber int
	heartBeatAttempCount int
	heartBeatAttempDuration time.Duration
	heartBeatCheckDuration time.Duration
}

func Default() *RegistryService {
	return New(3, 3, time.Second, 30 * time.Second)
}

func New(heartBeatWorkerNumber, heartBeatAttempCount int, heartBeatAttempDuration, heartBeatCheckDuration time.Duration) *RegistryService {
	return &RegistryService{
		serviceInfos: newServiceTable(),
		heartBeatWorkerNumber: heartBeatWorkerNumber,
		heartBeatAttempCount: heartBeatAttempCount,
		heartBeatAttempDuration: heartBeatAttempDuration,
		heartBeatCheckDuration: heartBeatCheckDuration,
	}
}

func (s *RegistryService) Run() error {
	go s.heartBeat()

	http.HandleFunc("/services", func(w http.ResponseWriter, r *http.Request) {
		statusCode := http.StatusOK
		switch r.Method {
		case http.MethodPost:
			registration, err := buildRegistration(r.Body)
			if err != nil {
				statusCode = http.StatusInternalServerError
				goto END
			}

			err = s.regist(registration)
			if err != nil {
				statusCode = http.StatusInternalServerError
				goto END
			}

			serviceInfos := s.serviceInfos.buildRequiredServiceInfos(registration)
			data, err := json.Marshal(&serviceInfos)
			if err != nil {
				statusCode = http.StatusInternalServerError
				goto END
			}
			defer w.Write(data)

		case http.MethodDelete:
			registration, err := buildRegistration(r.Body)
			if err != nil {
				statusCode = http.StatusInternalServerError
				goto END
			}

			s.unregist(registration)
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
	channel := make(chan *Registration, 1)
	for i := 0; i < s.heartBeatWorkerNumber; i++ {
		go func() {
			for reg := range channel {
				for j := 0; j < s.heartBeatAttempCount; j++ {
					resp, err := http.Get("http://" + reg.ServiceAddr + "/heart-beat")
					if err == nil && resp.StatusCode == http.StatusOK {
						goto NEXT
					}
					time.Sleep(s.heartBeatAttempDuration)
				}

				s.unregist(reg)

				NEXT:
			}
		}()
	}

	for {
		s.serviceInfos.lock.RLock()
		for _, registrations := range s.serviceInfos.serviceInfos {
			for i := len(registrations) - 1; i >= 0; i-- {
				channel <- registrations[i]
			}
		}
		s.serviceInfos.lock.RUnlock()
		time.Sleep(s.heartBeatCheckDuration)
	}
}

func (s *RegistryService) regist(registration *Registration) error {
	s.serviceInfos.add(registration)
	return s.serviceInfos.notify(http.MethodPost, registration)
}

func (s *RegistryService) unregist(registration *Registration) error {
	s.serviceInfos.remove(registration)
	return s.serviceInfos.notify(http.MethodDelete, registration)
}