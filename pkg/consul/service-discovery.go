package consul

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sync"
	"time"
	"github.com/hashicorp/consul/api"
)

type ServiceDiscovery interface {
	DiscoverService() (*api.CatalogService, error)
	CallAPI(service *api.CatalogService, endpoint, method string, body []byte, headers map[string]string) (string, error)
}

// serviceDiscovery - Struct to hold the Consul client and service name.
type serviceDiscovery struct {
	consulClient *api.Client
	serviceName  string
	once         sync.Once
}

// serviceDiscoveryMap - A map to store serviceDiscovery instances for each service name.
var serviceDiscoveryMap = make(map[string]*serviceDiscovery)
var mapMutex sync.Mutex

// NewServiceDiscovery - Constructor to initialize the serviceDiscovery with Consul client and service name.
func NewServiceDiscovery(client *api.Client, serviceName string) (*serviceDiscovery, error) {
	// Lock the map to avoid race condition while checking or inserting
	mapMutex.Lock()
	defer mapMutex.Unlock()

	// Check if the instance already exists in the map
	if sd, exists := serviceDiscoveryMap[serviceName]; exists {
		return sd, nil // Return existing instance
	}

	if client == nil {
		return nil, fmt.Errorf("error while creating Consul client")
	}

	sd := &serviceDiscovery{consulClient: client, serviceName: serviceName}

	// Use sync.Once to ensure that the serviceDiscovery setup runs only once
	sd.once.Do(func() {
		// Store the instance in the map
		serviceDiscoveryMap[serviceName] = sd
	})

	return sd, nil
}

// DiscoverService - Function to discover a service from Consul.
func (sd *serviceDiscovery) DiscoverService() (*api.CatalogService, error) {
	// Query the service in Consul
	services, _, err := sd.consulClient.Catalog().Service(sd.serviceName, "", nil)
	if err != nil {
		return nil, fmt.Errorf("error fetching service: %v", err)
	}

	if len(services) == 0 {
		return nil, fmt.Errorf("service %s not found in Consul", sd.serviceName)
	}

	// Randomly select one service instance (can be enhanced for load balancing)
	rand.New(rand.NewSource(time.Now().Unix()))
	service := services[rand.Intn(len(services))]
	return service, nil
}

// CallAPI - Function to send an HTTP request to the discovered service (supports GET, PUT, PATCH, DELETE, POST, etc.).
func (sd *serviceDiscovery) CallAPI(service *api.CatalogService, endpoint, method string, body []byte, headers map[string]string) (string, error) {
	// Build the API URL using service address and port
	url := fmt.Sprintf("http://%s:%d%s", service.ServiceAddress, service.ServicePort, endpoint)
	// Create a new HTTP request
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers if any
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	return string(bodyBytes), nil
}
