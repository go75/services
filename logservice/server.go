package logservice

import (
	"io"
	"log"
	"net/http"
	"os"
)

type logService struct {
	destination string
	logger *log.Logger
}

func Init(destination string) {
	s := &logService{
		destination: destination,
	}
	s.logger = log.New(s, "Go:", log.Ltime | log.Lshortfile)
	s.register()
}

func (s *logService)Write(data []byte) (int, error) {
	file, err := os.OpenFile(s.destination, os.O_CREATE | os.O_APPEND | os.O_WRONLY, 0600)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	return file.Write(data)
}

func (s *logService)register() {
	http.HandleFunc("/log", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		data, err := io.ReadAll(r.Body)
		if err != nil || len(data) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		s.logger.Println(string(data))
	})
}
