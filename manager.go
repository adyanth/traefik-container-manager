package traefik_container_manager

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const defaultTimeoutSeconds = 60

var netClient = &http.Client{
	Timeout: time.Second * 2,
}

// Config the plugin configuration
type Config struct {
	Name       string
	ServiceUrl string
	Timeout    uint64
}

// CreateConfig creates a config with its default values
func CreateConfig() *Config {
	return &Config{
		Timeout: defaultTimeoutSeconds,
	}
}

// Manager holds the request for the container
type Manager struct {
	name    string
	next    http.Handler
	request string
}

func buildRequest(url string, name string, timeout uint64) (string, error) {
	// TODO: Check url validity
	request := fmt.Sprintf("%s?name=%s&timeout=%d", url, name, timeout)
	fmt.Println(timeout)
	return request, nil
}

// New function creates the configuration
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {

	if len(config.Name) == 0 {
		return nil, fmt.Errorf("name cannot be null")
	}

	request, err := buildRequest(config.ServiceUrl, config.Name, config.Timeout)

	if err != nil {
		return nil, fmt.Errorf("error while building request")
	}

	return &Manager{
		next:    next,
		name:    name,
		request: request,
	}, nil
}

// ServeHTTP retrieve the service status
func (e *Manager) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	status, err := getServiceStatus(e.request)
	for err == nil && status == "starting" {
		status, err = getServiceStatus(e.request)
	}

	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(err.Error()))
	}

	if status == "started" {
		// Service started forward request
		e.next.ServeHTTP(rw, req)

	} else {
		// Error
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte("Unexpected status answer from Manager service"))
	}
}

func getServiceStatus(request string) (string, error) {

	// This request wakes up the service if he's scaled to 0
	resp, err := netClient.Get(request)
	if err != nil {
		return "error", err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "parsing error", err
	}

	return strings.TrimSuffix(string(body), "\n"), nil
}
