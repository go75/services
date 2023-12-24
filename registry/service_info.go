package registry

import (
	"encoding/json"
	"io"
)

type ServiceInfo struct {
	Name             string
	Addr             string
	RequiredServices []string
}

func buildServiceInfo(reader io.ReadCloser) (*ServiceInfo, error) {
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	serviceInfo := new(ServiceInfo)
	err = json.Unmarshal(data, serviceInfo)
	if err != nil {
		return nil, err
	}

	return serviceInfo, nil
}