package consul

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"event-service/config"
	"event-service/pkg/zap"
	"strconv"
	"time"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
)

const (
	serviceName = "event-service"
	ttl         = time.Second * 15
	checkId     = "event-service-health-check"
)

var (
	serviceId     = fmt.Sprintf("%s-%d", serviceName, rand.Intn(100))
	defaultConfig *api.Config
)

type Client interface {
	Connect() *api.Client
	Deregister()
}

type service struct {
	client *api.Client
	log    zap.Logger
	cfg    *config.Config
}

func NewConsulConn(log zap.Logger, cfg *config.Config) *service {
	consulHost := cfg.Consul.Host
	if consulHost == "" {
		// Fallback to localhost if the host is not set in the config
		consulHost = "localhost"
	}

	defaultConfig = &api.Config{
		Address: fmt.Sprintf("%s:%s", consulHost, cfg.Consul.Port), // Consul server address
		HttpClient: &http.Client{
			Timeout: 30 * time.Second, // Increase timeout to 30 seconds
		},
	}

	// Consul client setup
	client, err := api.NewClient(defaultConfig)

	if err != nil {
		log.Fatalf("Failed to create Consul client: %v", err)
	}

	return &service{
		client: client, // Store the client instance
		log:    log,
		cfg:    cfg,
	}
}

func (c *service) Connect() *api.Client {
	c.setupConsul()
	go c.updateHealthCheck()

	return c.client
}

func (c *service) Deregister() {
	// Deregister service
	err := c.client.Agent().ServiceDeregister(serviceId)
	if err != nil {
		c.log.Fatalf("Failed to deregister service: %v", err)
	}
}

func (c *service) updateHealthCheck() {
	ticker := time.NewTicker(time.Second * 5)

	for {
		err := c.client.Agent().UpdateTTL(checkId, "online", api.HealthPassing)
		if err != nil {
			log.Fatalf("Failed to check AgentHealthService: %v", err)
		}
		<-ticker.C
	}
}

func (c *service) setupConsul() {
	hostname := c.cfg.Registry.Host
	port, _ := strconv.Atoi(c.cfg.App.API.Rest.Port)

	// Health check (optional but recommended)
	check := &api.AgentServiceCheck{
		DeregisterCriticalServiceAfter: ttl.String(),
		TTL:                            ttl.String(),
		CheckID:                        checkId,
	}

	// Service registration
	registration := &api.AgentServiceRegistration{
		ID:      serviceId,   // Unique service ID
		Name:    serviceName, // Service name
		Port:    port,        // Service port
		Address: hostname,    // Service address
		Tags:    []string{"go", "chat-service"},
		Check:   check,
	}

	query := map[string]any{
		"type":        "service",
		"service":     serviceName,
		"passingonly": true,
	}

	plan, err := watch.Parse(query)
	if err != nil {
		c.log.Fatalf("Failed to watch for changes: %v", err)
	}

	plan.HybridHandler = func(index watch.BlockingParamVal, result interface{}) {
		switch msg := result.(type) {
		case []*api.ServiceEntry:
			for _, entry := range msg {
				c.log.Infof("new member <%s> joined, node <%s>", entry.Service.Service, entry.Node.Node)
			}
		default:
			c.log.Infof("Unexpected result type: %T", msg)
		}
	}

	go func() {
		_ = plan.RunWithConfig(fmt.Sprintf("%s:%s", c.cfg.Consul.Host, c.cfg.Consul.Port), api.DefaultConfig())
	}()

	err = c.client.Agent().ServiceRegister(registration)
	if err != nil {
		c.log.DPanic(err)
		c.log.Printf("Failed to register service: %s:%v ", hostname, port)
		c.log.Fatalf("Failed to register health check: %v", err)
	}

	c.log.Printf("successfully register service: %s:%v", hostname, port)
}
