package traefik_container_manager

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultTimeoutSeconds = 60 * 5             // 5 minutes
const defaultServiceUrl = "http://manager:10000" // Default URL is a container called manager serving on port 10000

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
		ServiceUrl: defaultServiceUrl,
		Timeout:    defaultTimeoutSeconds,
	}
}

// Manager holds the request for the container
type Manager struct {
	name       string
	next       http.Handler
	request    string
	serviceUrl string
	timeout    uint64
}

func buildRequest(baseUrl string, name string, timeout uint64, host, path string) (string, error) {
	// TODO: Check url validity
	request := fmt.Sprintf("%s?name=%s&timeout=%d&host=%s&path=%s", baseUrl, name, timeout, url.QueryEscape(host), url.QueryEscape(path))
	fmt.Println(timeout)
	return request, nil
}

// New function creates the configuration
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {

	if len(config.Name) == 0 {
		return nil, fmt.Errorf("name cannot be null")
	}

	request, err := buildRequest(config.ServiceUrl, config.Name, config.Timeout, "", "")

	if err != nil {
		return nil, fmt.Errorf("error while building request")
	}

	return &Manager{
		next:       next,
		name:       config.Name,
		request:    request,
		serviceUrl: config.ServiceUrl,
		timeout:    config.Timeout,
	}, nil
}

// ServeHTTP retrieve the service status
func (e *Manager) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if e.name == "generic-container-manager" {
		e.request, _ = buildRequest(e.serviceUrl, e.name, e.timeout, req.URL.Host, req.URL.Path)
		fmt.Println("Request set to ", e.request)
	}
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

	// This request starts up the service if it is stopped
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
