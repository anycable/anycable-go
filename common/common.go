// Package common contains struts and interfaces shared between multiple components
package common

import (
	"net"
	"net/http"
)

// SessionEnv represents the underlying HTTP connection data:
// URL parts and request headers
type SessionEnv struct {
	// Deprecated, to be removed in v1.1
	URL        string
	Path       string
	Query      string
	Host       string
	Port       string
	Scheme     string
	Origin     string
	RemoteAddr string
	Cookies    string
	Headers    *map[string]string
}

// SessionEnvFromRequest builds a SessionEnv struct from the http request and the list of required
// headers
func SessionEnvFromRequest(req *http.Request) *SessionEnv {
	remoteAddr, _, _ := net.SplitHostPort(req.RemoteAddr)

	return &SessionEnv{
		URL:        req.URL.String(),
		Path:       req.URL.Path,
		Query:      req.URL.RawQuery,
		Host:       req.URL.Hostname(),
		Port:       req.URL.Port(),
		Origin:     req.Header.Get("origin"),
		Cookies:    req.Header.Get("cookies"),
		Scheme:     req.URL.Scheme,
		RemoteAddr: remoteAddr,
	}
}

// CommandResult is a result of performing controller action,
// which contains informations about streams to subscribe,
// messages to sent and broadcast.
// It's a communication "protocol" between a node and a controller.
type CommandResult struct {
	Streams        []string
	StopAllStreams bool
	Transmissions  []string
	Disconnect     bool
	Broadcasts     []*StreamMessage
}

// Message represents incoming client message
type Message struct {
	Command    string `json:"command"`
	Identifier string `json:"identifier"`
	Data       string `json:"data"`
}

// StreamMessage represents a pub/sub message to be sent to stream
type StreamMessage struct {
	Stream string `json:"stream"`
	Data   string `json:"data"`
}
