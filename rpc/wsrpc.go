package rpc

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	pb "github.com/anycable/anycable-go/protos"
	"github.com/anycable/anycable-go/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type WSClientHelper struct {
	service *WSService
}

func NewWSClientHelper(s *WSService) *WSClientHelper {
	return &WSClientHelper{service: s}
}

func (h *WSClientHelper) Ready() error {
	return nil
}

func (h *WSClientHelper) SupportsActiveConns() bool {
	return false
}

func (h *WSClientHelper) ActiveConns() int {
	return 0
}

func (h *WSClientHelper) Close() {
	h.service.Close()
}

//go:generate mockery --name WSClient --output "../mocks" --outpkg mocks
type WSClient interface {
	Invoke(ctx context.Context, command string, payload []byte, meta *map[string]string) ([]byte, int, error)
	Shutdown(ctx context.Context) error
}

type WSService struct {
	conf   *Config
	client WSClient
}

func NewWSDialer(c *Config, client WSClient) (Dialer, error) {
	service, err := NewWSService(c, client)

	if err != nil {
		return nil, err
	}

	helper := NewWSClientHelper(service)

	return NewInprocessServiceDialer(service, helper), nil
}

func NewWSService(c *Config, client WSClient) (*WSService, error) {
	return &WSService{conf: c, client: client}, nil
}

func (s *WSService) Connect(ctx context.Context, r *pb.ConnectionRequest) (*pb.ConnectionResponse, error) {
	rawResponse, err := s.performRequest(ctx, "connect", utils.ToJSON(r))

	if err != nil {
		return nil, err
	}

	var response pb.ConnectionResponse

	err = json.Unmarshal(rawResponse, &response)

	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (s *WSService) Disconnect(ctx context.Context, r *pb.DisconnectRequest) (*pb.DisconnectResponse, error) {
	rawResponse, err := s.performRequest(ctx, "disconnect", utils.ToJSON(r))

	if err != nil {
		return nil, err
	}

	var response pb.DisconnectResponse

	err = json.Unmarshal(rawResponse, &response)

	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (s *WSService) Command(ctx context.Context, r *pb.CommandMessage) (*pb.CommandResponse, error) {
	rawResponse, err := s.performRequest(ctx, "command", utils.ToJSON(r))

	if err != nil {
		return nil, err
	}

	var response pb.CommandResponse

	err = json.Unmarshal(rawResponse, &response)

	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (s *WSService) Close() {
	s.client.Shutdown(context.Background()) //nolint:errcheck
}

func (s *WSService) performRequest(ctx context.Context, command string, payload []byte) ([]byte, error) {
	// We use timeouts to detect request queueing at the WS RPC side and report ResourceExhausted errors
	// (so adaptive concurrency control can be applied)
	ctx, cancel := context.WithTimeout(ctx, time.Duration(s.conf.RequestTimeout)*time.Millisecond)
	defer cancel()

	var meta *map[string]string

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		m := make(map[string]string)
		meta = &m
		// Set headers from metadata
		for k, v := range md {
			m[k] = v[0]
		}
	}

	result, code, err := s.client.Invoke(ctx, command, payload, meta)

	if err != nil {
		if ctx.Err() != nil {
			return nil, status.Error(codes.DeadlineExceeded, "request timeout")
		}

		return nil, status.Error(codes.Unavailable, err.Error())
	}

	if code == http.StatusUnauthorized {
		return nil, status.Error(codes.Unauthenticated, "http returned 401")
	}

	if code == http.StatusUnprocessableEntity {
		return nil, status.Error(codes.InvalidArgument, "unprocessable entity")
	}

	if code != http.StatusOK {
		return nil, status.Error(codes.Unknown, "internal error")
	}

	return result, nil
}
