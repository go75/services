package gatewayservice

import (
	"io"
	"net/http"
	"services/registry"
	"strings"
)

func Init() {
	register()
}

func register() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.SplitN(r.URL.Path, "/", 3)
		if len(parts) < 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		addr := registry.Get(parts[1])
		if addr == "" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		req, err := http.NewRequest(r.Method, "http://" + addr + r.URL.String(), r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		for k, v := range resp.Header {
			w.Header()[k] = v
		}
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	})
}