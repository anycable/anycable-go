package node

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/anycable/anycable-go/common"
	"github.com/anycable/anycable-go/encoders"
	"github.com/anycable/anycable-go/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAuthenticate(t *testing.T) {
	node := NewMockNode()
	go node.hub.Run()
	defer node.hub.Shutdown()

	t.Run("Successful authentication", func(t *testing.T) {
		session := NewMockSessionWithEnv("1", &node, "/cable", &map[string]string{"id": "test_id"})
		_, err := node.Authenticate(session)
		defer node.hub.RemoveSession(session)

		assert.Nil(t, err, "Error must be nil")
		assert.Equal(t, true, session.Connected, "Session must be marked as connected")
		assert.Equalf(t, "test_id", session.GetIdentifiers(), "Identifiers must be equal to %s", "test_id")

		msg, err := session.conn.Read()
		assert.Nil(t, err)

		assert.Equalf(t, []byte("welcome"), msg, "Sent message is invalid: %s", msg)

		assert.Equal(t, 1, node.hub.Size())
	})

	t.Run("Failed authentication", func(t *testing.T) {
		session := NewMockSessionWithEnv("1", &node, "/failure", &map[string]string{"id": "test_id"})

		_, err := node.Authenticate(session)

		assert.Nil(t, err, "Error must be nil")

		msg, err := session.conn.Read()
		assert.Nil(t, err)

		assert.Equalf(t, []byte("unauthorized"), msg, "Sent message is invalid: %s", msg)
		assert.Equal(t, 0, node.hub.Size())
	})

	t.Run("Error during authentication", func(t *testing.T) {
		session := NewMockSessionWithEnv("1", &node, "/error", &map[string]string{"id": "test_id"})

		_, err := node.Authenticate(session)

		assert.NotNil(t, err, "Error must not be nil")
		assert.Equal(t, 0, node.hub.Size())
	})

	t.Run("With connection state", func(t *testing.T) {
		session := NewMockSessionWithEnv("1", &node, "/cable", &map[string]string{"x-session-test": "my_session", "id": "session_id"})
		defer node.hub.RemoveSession(session)

		_, err := node.Authenticate(session)

		assert.Nil(t, err, "Error must be nil")
		assert.Equal(t, true, session.Connected, "Session must be marked as connected")

		assert.Len(t, *session.env.ConnectionState, 1)
		assert.Equal(t, "my_session", (*session.env.ConnectionState)["_s_"])

		assert.Equal(t, 1, node.hub.Size())
	})
}

func TestSubscribe(t *testing.T) {
	node := NewMockNode()
	go node.hub.Run()
	defer node.hub.Shutdown()

	session := NewMockSession("14", &node)

	t.Run("Successful subscription", func(t *testing.T) {
		_, err := node.Subscribe(session, &common.Message{Identifier: "test_channel"})
		assert.Nil(t, err, "Error must be nil")

		// Adds subscription to session
		assert.Truef(t, session.subscriptions.HasChannel("test_channel"), "Session subscription must be set")

		msg, err := session.conn.Read()
		assert.Nil(t, err)

		assert.Equalf(t, []byte("14"), msg, "Sent message is invalid: %s", msg)
	})

	t.Run("Subscription with a stream", func(t *testing.T) {
		node.hub.AddSession(session)
		defer node.hub.RemoveSession(session)

		_, err := node.Subscribe(session, &common.Message{Identifier: "with_stream"})
		assert.Nil(t, err, "Error must be nil")

		// Adds subscription and stream to session
		assert.Truef(t, session.subscriptions.HasChannel("with_stream"), "Session subsription must be set")
		assert.Equal(t, []string{"stream"}, session.subscriptions.StreamsFor("with_stream"))

		msg, err := session.conn.Read()
		assert.Nil(t, err)

		assert.Equalf(t, "14", string(msg), "Sent message is invalid: %s", msg)

		// Make sure session is subscribed
		node.hub.BroadcastMessage(&common.StreamMessage{Stream: "stream", Data: "41"})

		msg, err = session.conn.Read()
		assert.Nil(t, err)

		assert.Equalf(t, "{\"identifier\":\"with_stream\",\"message\":41}", string(msg), "Broadcasted message is invalid: %s", msg)
	})

	t.Run("Error during subscription", func(t *testing.T) {
		_, err := node.Subscribe(session, &common.Message{Identifier: "error"})
		assert.NotNil(t, err, "Error must not be nil")
	})

	t.Run("Rejected subscription", func(t *testing.T) {
		res, err := node.Subscribe(session, &common.Message{Identifier: "failure"})

		assert.Equal(t, common.FAILURE, res.Status)
		assert.Nil(t, err, "Error must be nil")
	})
}

func TestUnsubscribe(t *testing.T) {
	node := NewMockNode()
	go node.hub.Run()
	defer node.hub.Shutdown()

	session := NewMockSession("14", &node)

	t.Run("Successful unsubscribe", func(t *testing.T) {
		session.subscriptions.AddChannel("test_channel")
		node.hub.SubscribeSession("14", "streamo", "test_channel")

		_, err := node.Unsubscribe(session, &common.Message{Identifier: "test_channel"})
		assert.Nil(t, err, "Error must be nil")

		// Removes subscription from session
		assert.Falsef(t, session.subscriptions.HasChannel("test_channel"), "Shouldn't contain test_channel")

		msg, err := session.conn.Read()
		assert.Nil(t, err)

		assert.Equalf(t, []byte("14"), msg, "Sent message is invalid: %s", msg)

		node.hub.BroadcastMessage(&common.StreamMessage{Stream: "streamo", Data: "41"})

		msg, err = session.conn.Read()
		assert.Nil(t, msg)
		assert.Error(t, err, "Session hasn't received any messages")
	})

	t.Run("Error during unsubscription", func(t *testing.T) {
		session.subscriptions.AddChannel("failure")

		_, err := node.Unsubscribe(session, &common.Message{Identifier: "failure"})
		assert.NotNil(t, err, "Error must not be nil")
	})
}

func TestPerform(t *testing.T) {
	node := NewMockNode()
	go node.hub.Run()
	defer node.hub.Shutdown()

	session := NewMockSession("14", &node)
	node.hub.AddSession(session)

	session.subscriptions.AddChannel("test_channel")

	t.Run("Successful perform", func(t *testing.T) {
		_, err := node.Perform(session, &common.Message{Identifier: "test_channel", Data: "action"})
		assert.Nil(t, err)

		msg, err := session.conn.Read()
		assert.Nil(t, err)

		assert.Equalf(t, []byte("action"), msg, "Sent message is invalid: %s", msg)
	})

	t.Run("With connection state", func(t *testing.T) {
		_, err := node.Perform(session, &common.Message{Identifier: "test_channel", Data: "session"})
		assert.Nil(t, err)

		_, err = session.conn.Read()
		assert.Nil(t, err)

		assert.Len(t, *session.env.ConnectionState, 1)
		assert.Equal(t, "performed", (*session.env.ConnectionState)["_s_"])
	})

	t.Run("Error during perform", func(t *testing.T) {
		session.subscriptions.AddChannel("failure")

		_, err := node.Perform(session, &common.Message{Identifier: "failure", Data: "test"})
		assert.NotNil(t, err, "Error must not be nil")
	})

	t.Run("With stopped streams", func(t *testing.T) {
		session.subscriptions.AddChannelStream("test_channel", "stop_stream")
		node.hub.SubscribeSession("14", "stop_stream", "test_channel")

		node.hub.BroadcastMessage(&common.StreamMessage{Stream: "stop_stream", Data: "40"})

		msg, _ := session.conn.Read()
		assert.NotNil(t, msg)

		_, err := node.Perform(session, &common.Message{Identifier: "test_channel", Data: "stop_stream"})
		assert.Nil(t, err)

		assert.Empty(t, session.subscriptions.StreamsFor("test_channel"))

		_, err = session.conn.Read()
		assert.Nil(t, err)

		node.hub.BroadcastMessage(&common.StreamMessage{Stream: "stop_stream", Data: "41"})

		msg, err = session.conn.Read()
		assert.Nil(t, msg)
		assert.Error(t, err, "Session hasn't received any messages")
	})

	t.Run("With channel state", func(t *testing.T) {
		assert.Len(t, *session.env.ChannelStates, 0)

		_, err := node.Perform(session, &common.Message{Identifier: "test_channel", Data: "channel_state"})
		assert.Nil(t, err)

		_, err = session.conn.Read()
		assert.Nil(t, err)

		assert.Len(t, *session.env.ChannelStates, 1)
		assert.Len(t, (*session.env.ChannelStates)["test_channel"], 1)
		assert.Equal(t, "performed", (*session.env.ChannelStates)["test_channel"]["_c_"])
	})
}

func TestDisconnect(t *testing.T) {
	node := NewMockNode()
	go node.hub.Run()
	defer node.hub.Shutdown()

	session := NewMockSession("14", &node)

	assert.Nil(t, node.Disconnect(session))

	assert.Equal(t, node.disconnector.Size(), 1, "Expected disconnect to have 1 task in a queue")

	task := <-node.disconnector.(*DisconnectQueue).disconnect
	assert.Equal(t, session, task, "Expected to disconnect session")

	assert.Equal(t, node.hub.Size(), 0)
}

func TestHistory(t *testing.T) {
	node := NewMockNode()

	broker := &mocks.Broker{}
	node.SetBroker(broker)

	go node.hub.Run()
	defer node.hub.Shutdown()

	session := NewMockSession("14", &node)

	session.subscriptions.AddChannel("test_channel")
	session.subscriptions.AddChannelStream("test_channel", "streamo")
	session.subscriptions.AddChannelStream("test_channel", "emptissimo")

	stream := []common.StreamMessage{
		{
			Stream: "streamo",
			Data:   "ciao",
			Offset: 22,
			Epoch:  "test",
		},
		{
			Stream: "streamo",
			Data:   "buona sera",
			Offset: 23,
			Epoch:  "test",
		},
	}

	ts := int(time.Now().Unix())

	t.Run("Successful history with only Since", func(t *testing.T) {
		broker.ExpectedCalls = []*mock.Call{}

		broker.
			On("HistorySince", "streamo", ts).
			Return(stream, nil)
		broker.
			On("HistorySince", "emptissimo", ts).
			Return(nil, nil)

		err := node.History(
			session,
			&common.Message{
				Identifier: "test_channel",
				History: common.HistoryRequest{
					Since: ts,
				},
			},
		)
		require.NoError(t, err)

		history := []string{
			"{\"identifier\":\"test_channel\",\"message\":\"ciao\",\"stream_id\":\"streamo\",\"epoch\":\"test\",\"offset\":22}",
			"{\"identifier\":\"test_channel\",\"message\":\"buona sera\",\"stream_id\":\"streamo\",\"epoch\":\"test\",\"offset\":23}",
		}

		for _, msg := range history {
			received, herr := session.conn.Read()
			require.NoError(t, herr)

			require.Equalf(
				t,
				msg,
				string(received),
				"Sent message is invalid: %s", received,
			)
		}

		_, err = session.conn.Read()
		require.Error(t, err)
	})

	t.Run("Successful history with Since and Offset", func(t *testing.T) {
		broker.ExpectedCalls = []*mock.Call{}

		broker.
			On("HistoryFrom", "streamo", "test", uint64(20)).
			Return(stream, nil)
		broker.
			On("HistorySince", "emptissimo", ts).
			Return([]common.StreamMessage{{
				Stream: "emptissimo",
				Data:   "zer0",
				Offset: 2,
				Epoch:  "test_zero",
			}}, nil)

		err := node.History(
			session,
			&common.Message{
				Identifier: "test_channel",
				History: common.HistoryRequest{
					Since: ts,
					Streams: map[string]common.HistoryPosition{
						"streamo": {Epoch: "test", Offset: 20},
					},
				},
			},
		)
		require.NoError(t, err)

		history := []string{
			"{\"identifier\":\"test_channel\",\"message\":\"ciao\",\"stream_id\":\"streamo\",\"epoch\":\"test\",\"offset\":22}",
			"{\"identifier\":\"test_channel\",\"message\":\"buona sera\",\"stream_id\":\"streamo\",\"epoch\":\"test\",\"offset\":23}",
			"{\"identifier\":\"test_channel\",\"message\":\"zer0\",\"stream_id\":\"emptissimo\",\"epoch\":\"test_zero\",\"offset\":2}",
		}

		// The order of streams is non-deterministic, so
		// we're collecting messages first and checking for inclusion later
		received := []string{}

		for range history {
			data, herr := session.conn.Read()
			require.NoError(t, herr)

			received = append(received, string(data))
		}

		for _, msg := range history {
			require.Contains(
				t,
				received,
				msg,
			)
		}

		_, err = session.conn.Read()
		require.Error(t, err)
	})

	t.Run("Fetching history with Subscribe", func(t *testing.T) {
		broker.ExpectedCalls = []*mock.Call{}

		broker.
			On("HistoryFrom", "streamo", "test", uint64(20)).
			Return(stream, nil)
		broker.
			On("HistorySince", "s1", ts).
			Return([]common.StreamMessage{{
				Stream: "s1",
				Data:   "{\"foo\":\"bar\"}",
				Offset: 10,
				Epoch:  "test",
			}}, nil)
		broker.
			On("Subscribe", "stream").
			Return("s1")

		_, err := node.Subscribe(
			session,
			&common.Message{
				Identifier: "with_stream",
				History: common.HistoryRequest{
					Since: ts,
					Streams: map[string]common.HistoryPosition{
						"streamo": {Epoch: "test", Offset: 20},
					},
				},
			},
		)
		require.NoError(t, err)

		msg, err := session.conn.Read()
		require.NoError(t, err)

		require.Equalf(t, "14", string(msg), "Sent message is invalid: %s", msg)

		history := []string{
			"{\"identifier\":\"with_stream\",\"message\":{\"foo\":\"bar\"},\"stream_id\":\"s1\",\"epoch\":\"test\",\"offset\":10}",
		}

		for _, msg := range history {
			received, herr := session.conn.Read()
			require.NoError(t, herr)

			require.Equalf(
				t,
				msg,
				string(received),
				"Sent message is invalid: %s", received,
			)
		}

		_, err = session.conn.Read()
		require.Error(t, err)
	})

	t.Run("Error retrieving history", func(t *testing.T) {
		broker.ExpectedCalls = []*mock.Call{}

		broker.
			On("HistorySince", "streamo", ts).
			Return(nil, errors.New("Couldn't restore history"))
		broker.
			On("HistorySince", "emptissimo", ts).
			Return(stream, nil)

		err := node.History(
			session,
			&common.Message{
				Identifier: "test_channel",
				History: common.HistoryRequest{
					Since: ts,
				},
			},
		)

		assert.Error(t, err, "Couldn't restore history")
	})
}

func TestHandlePubSub(t *testing.T) {
	node := NewMockNode()

	go node.hub.Run()
	defer node.hub.Shutdown()

	session := NewMockSession("14", &node)
	session2 := NewMockSession("15", &node)

	node.hub.AddSession(session)
	node.hub.SubscribeSession("14", "test", "test_channel")

	node.hub.AddSession(session2)
	node.hub.SubscribeSession("15", "test", "test_channel")

	node.HandlePubSub([]byte("{\"stream\":\"test\",\"data\":\"\\\"abc123\\\"\"}"))

	expected := "{\"identifier\":\"test_channel\",\"message\":\"abc123\"}"

	msg, err := session.conn.Read()
	assert.Nil(t, err)
	assert.Equalf(t, expected, string(msg), "Expected to receive %s but got %s", expected, string(msg))

	msg2, err := session2.conn.Read()
	assert.Nil(t, err)
	assert.Equalf(t, expected, string(msg2), "Expected to receive %s but got %s", expected, string(msg2))
}

func TestHandlePubSubWithCommand(t *testing.T) {
	node := NewMockNode()

	go node.hub.Run()
	defer node.hub.Shutdown()

	session := NewMockSession("14", &node)
	node.hub.AddSession(session)

	node.HandlePubSub([]byte("{\"command\":\"disconnect\",\"payload\":{\"identifier\":\"14\",\"reconnect\":false}}"))

	expected := string(toJSON(common.NewDisconnectMessage("remote", false)))

	msg, err := session.conn.Read()
	assert.Nil(t, err)
	assert.Equalf(t, expected, string(msg), "Expected to receive %s but got %s", expected, string(msg))
	assert.True(t, session.closed)
}

func TestLookupSession(t *testing.T) {
	node := NewMockNode()

	go node.hub.Run()
	defer node.hub.Shutdown()

	assert.Nil(t, node.LookupSession("{\"foo\":\"bar\"}"))

	session := NewMockSession("14", &node)
	session.SetIdentifiers("{\"foo\":\"bar\"}")
	node.hub.AddSession(session)

	assert.Equal(t, session, node.LookupSession("{\"foo\":\"bar\"}"))
}

func toJSON(msg encoders.EncodedMessage) []byte {
	b, err := json.Marshal(&msg)
	if err != nil {
		panic("Failed to build JSON 😲")
	}

	return b
}
