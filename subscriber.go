package main

import (
	"encoding/json"

	"github.com/garyburd/redigo/redis"
	"github.com/soveran/redisurl"
)

type Subscriber struct {
	host    string
	channel string
}

func (s *Subscriber) run() {
	c, err := redisurl.ConnectToURL(s.host)

	if err != nil {
		app.Logger.Criticalf("failed to subscribe to Redis: %v", err)
		return
	}

	psc := redis.PubSubConn{Conn: c}
	psc.Subscribe(s.channel)
	for {
		switch v := psc.Receive().(type) {
		case redis.Message:
			app.Logger.Debugf("[Redis] channel %s: message: %s\n", v.Channel, v.Data)
			msg := &StreamMessage{}
			if err := json.Unmarshal(v.Data, &msg); err != nil {
				app.Logger.Debugf("Unknown message: %s", v.Data)
			} else {
				app.Logger.Debugf("Broadcast %v", msg)
				hub.stream_broadcast <- msg
			}
		case redis.Subscription:
			app.Logger.Debugf("%s: %s %d\n", v.Channel, v.Kind, v.Count)
		case error:
			break
		}
	}
}
