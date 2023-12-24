package registry

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type ServiceInfo struct {
	Name string
	Addr string
	RequiredServices []string
}

func buildRegistration(reader io.ReadCloser) (*ServiceInfo, error) {
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

func buildServiceInfo(reader io.ReadCloser) ([]string, error) {
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	parts := strings.SplitN(string(data), " ", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("Parse service failed with length %d", len(parts))
	}

	return parts, nil
}