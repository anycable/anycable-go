package ext

import (
	"net/url"
	"strconv"
	"time"

	"github.com/apex/log"
	"github.com/joomcode/errorx"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

const (
	natsServerStartTimeout = 5 * time.Second
)

// NATSServiceConfig represents NATS service configuration
type NATSServiceConfig struct {
	Enable        bool
	Debug         bool
	Trace         bool
	EnableLogging bool
	ServiceAddr   string
	ClusterAddr   string
	ClusterName   string
	Routes        []string
}

// NATSService represents NATS service
type NATSService struct {
	config NATSServiceConfig
	server *server.Server
}

// NATSServerLogEntry represents LoggerV2 decorator for nats server logger
type NATSServerLogEntry struct {
	*log.Entry
}

// Noticef is an alias for Infof
func (e *NATSServerLogEntry) Noticef(format string, v ...interface{}) {
	e.Infof(format, v...)
}

// Tracef is an alias for Debugf
func (e *NATSServerLogEntry) Tracef(format string, v ...interface{}) {
	e.Debugf(format, v...)
}

// NewNATSServiceConfig returns defaults for NATSServiceConfig
func NewNATSServiceConfig() NATSServiceConfig {
	return NATSServiceConfig{ServiceAddr: nats.DefaultURL, ClusterName: "anycable-cluster"}
}

// NewNATSService returns an instance of NATS service
func NewNATSService(c NATSServiceConfig) Service {
	return &NATSService{config: c}
}

// Start starts the service
func (s *NATSService) Start() error {
	var clusterOpts server.ClusterOpts
	var err error

	u, err := url.Parse(s.config.ServiceAddr)
	if err != nil {
		return errorx.Decorate(err, "Error parsing NATS service addr")
	}
	if u.Port() == "" {
		return errorx.IllegalArgument.New("Failed to parse NATS server URL, can not fetch port")
	}

	port, err := strconv.ParseInt(u.Port(), 10, 32)
	if err != nil {
		return errorx.Decorate(err, "Failed to parse NATS service port")
	}

	if s.config.ClusterAddr != "" {
		var clusterURL *url.URL
		var clusterPort int64

		clusterURL, err = url.Parse(s.config.ClusterAddr)
		if err != nil {
			return errorx.Decorate(err, "Failed to parse NATS cluster URL")
		}
		if clusterURL.Port() == "" {
			return errorx.IllegalArgument.New("Failed to parse NATS cluster port")
		}

		clusterPort, err = strconv.ParseInt(clusterURL.Port(), 10, 32)
		if err != nil {
			return errorx.Decorate(err, "Failed to parse NATS cluster port")
		}
		clusterOpts = server.ClusterOpts{
			Name: s.config.ClusterName,
			Host: clusterURL.Hostname(),
			Port: int(clusterPort),
		}
	}

	routes, err := s.getRoutes()
	if err != nil {
		return errorx.Decorate(err, "Failed to parse routes")
	}

	opts := &server.Options{
		Host:    u.Hostname(),
		Port:    int(port),
		Debug:   s.config.Debug,
		Trace:   s.config.Trace,
		Cluster: clusterOpts,
		Routes:  routes,
	}

	s.server, err = server.NewServer(opts)
	if err != nil {
		return errorx.Decorate(err, "Failed to start NATS server")
	}

	if s.config.EnableLogging || s.config.Debug {
		e := &NATSServerLogEntry{log.WithField("service", "nats")}
		s.server.SetLogger(e, s.config.Debug, s.config.Trace)
	}

	go s.server.Start()

	return nil
}

// WaitReady waits while NATS server is starting
func (s *NATSService) WaitReady() error {
	if s.server.ReadyForConnections(natsServerStartTimeout) {
		return nil
	}

	return errorx.TimeoutElapsed.New(
		"Failed to start NATS server within %d seconds", natsServerStartTimeout,
	)
}

// Shutdown shuts the NATS server down
func (s *NATSService) Shutdown() error {
	s.server.Shutdown()
	s.server.WaitForShutdown()
	return nil
}

// getRoutes transforms []string routes to []*url.URL routes
func (s *NATSService) getRoutes() ([]*url.URL, error) {
	if len(s.config.Routes) == 0 {
		return nil, nil
	}

	routes := make([]*url.URL, len(s.config.Routes))
	for i, r := range s.config.Routes {
		u, err := url.Parse(r)
		if err != nil {
			return nil, errorx.Decorate(err, "Error parsing route URL")
		}
		routes[i] = u
	}
	return routes, nil
}
