package rpc

import (
	"context"
	"net/http"
	"testing"

	"github.com/anycable/anycable-go/common"
	"github.com/anycable/anycable-go/mocks"
	"github.com/anycable/anycable-go/protocol"
	pb "github.com/anycable/anycable-go/protos"
	"github.com/anycable/anycable-go/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestWSService_Connect(t *testing.T) {
	t.Run("successful connection", func(t *testing.T) {
		client := &mocks.WSClient{}

		service, _ := NewWSService(&Config{RequestTimeout: 1000}, client)

		request := protocol.NewConnectMessage(
			common.NewSessionEnv("ws://anycable.io/cable", &map[string]string{"cookie": "foo=bar"}),
		)

		client.On("Invoke", mock.Anything, "connect", utils.ToJSON(request), mock.Anything).Return(
			[]byte(`{"status": 1, "transmissions": ["welcome"], "identifiers": "2024-Spb"}`),
			http.StatusOK,
			nil,
		)

		res, err := service.Connect(context.Background(), request)

		require.NoError(t, err)
		assert.Equal(t, pb.Status_SUCCESS, res.Status)
		assert.Equal(t, []string{"welcome"}, res.Transmissions)
		assert.Equal(t, "2024-Spb", res.Identifiers)
	})

	t.Run("unauthorized connection", func(t *testing.T) {
		client := &mocks.WSClient{}

		client.On("Invoke", mock.Anything, "connect", mock.Anything, mock.Anything).Return(
			nil,
			http.StatusUnauthorized,
			nil,
		)

		service, _ := NewWSService(&Config{RequestTimeout: 1000}, client)

		_, err := service.Connect(context.Background(), &pb.ConnectionRequest{})

		assert.Error(t, err)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})
}

func TestWSService_Command(t *testing.T) {
	t.Run("successful command", func(t *testing.T) {
		client := &mocks.WSClient{}

		service, _ := NewWSService(&Config{RequestTimeout: 1000}, client)

		request := &pb.CommandMessage{
			Command:    "subscribe",
			Identifier: "test_channel",
		}

		client.On("Invoke", mock.Anything, "command", utils.ToJSON(request), mock.Anything).Return(
			[]byte(`{"status": 1, "transmissions": ["acknowledged"], "streams": ["ok"]}`),
			http.StatusOK,
			nil,
		)

		res, err := service.Command(context.Background(), request)

		require.NoError(t, err)
		assert.Equal(t, pb.Status_SUCCESS, res.Status)
		assert.Equal(t, []string{"acknowledged"}, res.Transmissions)
		assert.Equal(t, []string{"ok"}, res.Streams)
	})
}

func TestWSService_Disconnect(t *testing.T) {
	t.Run("successful disconnect", func(t *testing.T) {
		client := &mocks.WSClient{}

		service, _ := NewWSService(&Config{RequestTimeout: 1000}, client)

		request := &pb.DisconnectRequest{
			Identifiers:   "test_session",
			Subscriptions: []string{"test_channel"},
		}

		client.On("Invoke", mock.Anything, "disconnect", utils.ToJSON(request), mock.Anything).Return(
			[]byte(`{"status": 1}`),
			http.StatusOK,
			nil,
		)

		res, err := service.Disconnect(context.Background(), request)

		require.NoError(t, err)
		assert.Equal(t, pb.Status_SUCCESS, res.Status)
	})
}
