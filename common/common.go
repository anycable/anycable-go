// Package common contains struts and interfaces shared between multiple components
package common

import (
	"encoding/json"
	"fmt"
)

// SessionEnv represents the underlying HTTP connection data:
// URL and request headers
type SessionEnv struct {
	URL             string
	Headers         *map[string]string
	ConnectionState *map[string]string
	ChannelStates   *map[string]map[string]string
}

// NewSessionEnv builds a new SessionEnv
func NewSessionEnv(url string, headers *map[string]string) *SessionEnv {
	state := make(map[string]string)
	channels := make(map[string]map[string]string)
	return &SessionEnv{
		URL:             url,
		Headers:         headers,
		ConnectionState: &state,
		ChannelStates:   &channels,
	}
}

// MergeConnectionState updates the current ConnectionState from the given map.
// If the value is an empty string then remove the key,
// otherswise add or rewrite.
func (st *SessionEnv) MergeConnectionState(other *map[string]string) {
	for k, v := range *other {
		if v == "" {
			delete(*st.ConnectionState, k)
		} else {
			(*st.ConnectionState)[k] = v
		}
	}
}

// MergeChannelState updates the current ChannelStates for the given identifier.
// If the value is an empty string then remove the key,
// otherswise add or rewrite.
func (st *SessionEnv) MergeChannelState(id string, other *map[string]string) {
	if _, ok := (*st.ChannelStates)[id]; !ok {
		(*st.ChannelStates)[id] = make(map[string]string)
	}

	for k, v := range *other {
		if v == "" {
			delete((*st.ChannelStates)[id], k)
		} else {
			(*st.ChannelStates)[id][k] = v
		}
	}
}

// CallResult contains shared RPC result fields
type CallResult struct {
	Transmissions []string
	Broadcasts    []*StreamMessage
	CState        map[string]string
	IState        map[string]string
}

// ConnectResult is a result of initializing a connection (calling a Connect method)
type ConnectResult struct {
	Identifier    string
	Transmissions []string
	Broadcasts    []*StreamMessage
	CState        map[string]string
	IState        map[string]string
}

// ToCallResult returns the corresponding CallResult
func (c *ConnectResult) ToCallResult() *CallResult {
	res := CallResult{Transmissions: c.Transmissions, Broadcasts: c.Broadcasts}
	if c.CState != nil {
		res.CState = c.CState
	}
	if c.IState != nil {
		res.IState = c.IState
	}
	return &res
}

// CommandResult is a result of performing controller action,
// which contains informations about streams to subscribe,
// messages to sent and broadcast.
// It's a communication "protocol" between a node and a controller.
type CommandResult struct {
	StopAllStreams bool
	Disconnect     bool
	Streams        []string
	StoppedStreams []string
	Transmissions  []string
	Broadcasts     []*StreamMessage
	CState         map[string]string
	IState         map[string]string
}

// ToCallResult returns the corresponding CallResult
func (c *CommandResult) ToCallResult() *CallResult {
	res := CallResult{Transmissions: c.Transmissions, Broadcasts: c.Broadcasts}
	if c.CState != nil {
		res.CState = c.CState
	}
	if c.IState != nil {
		res.IState = c.IState
	}
	return &res
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

// RemoteCommandMessage represents a pub/sub message with a remote command (e.g., disconnect)
type RemoteCommandMessage struct {
	Command string          `json:"command,omitempty"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// RemoteDisconnectMessage contains information required to disconnect a session
type RemoteDisconnectMessage struct {
	Identifier string `json:"identifier"`
	Reconnect  bool   `json:"reconnect"`
}

// PingMessage represents a server ping
type PingMessage struct {
	Type    string      `json:"type"`
	Message interface{} `json:"message"`
}

// DisconnectMessage represents a server disconnect message
type DisconnectMessage struct {
	Type      string `json:"type"`
	Reason    string `json:"reason"`
	Reconnect bool   `json:"reconnect"`
}

// Reply represents outgoing client message
type Reply struct {
	Type       string      `json:"type,omitempty"`
	Identifier string      `json:"identifier"`
	Message    interface{} `json:"message"`
}

// PubSubMessageFromJSON takes raw JSON byte array and return the corresponding struct
func PubSubMessageFromJSON(raw []byte) (interface{}, error) {
	smsg := StreamMessage{}

	if err := json.Unmarshal(raw, &smsg); err == nil {
		if smsg.Stream != "" {
			return smsg, nil
		}
	}

	rmsg := RemoteCommandMessage{}

	if err := json.Unmarshal(raw, &rmsg); err != nil {
		return nil, err
	}

	if rmsg.Command == "disconnect" {
		dmsg := RemoteDisconnectMessage{}

		if err := json.Unmarshal(rmsg.Payload, &dmsg); err != nil {
			return nil, err
		}

		return dmsg, nil
	}

	return nil, fmt.Errorf("Unknown message: %s", raw)
}
