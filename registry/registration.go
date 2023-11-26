package registry

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type Registration struct {
	ServiceName string
	ServiceAddr string
	RequiredServices []string
}

func buildRegistration(reader io.ReadCloser) (*Registration, error) {
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	registration := new(Registration)
	err = json.Unmarshal(data, registration)
	if err != nil {
		return nil, err
	}

	return registration, nil
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