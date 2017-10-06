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
		log.Criticalf("failed to subscribe to Redis: %v", err)
		return
	}

	psc := redis.PubSubConn{Conn: c}
	psc.Subscribe(s.channel)
	for {
		switch v := psc.Receive().(type) {
		case redis.Message:
			log.Debugf("[Redis] channel %s: message: %s\n", v.Channel, v.Data)
			msg := &StreamMessage{}
			if err := json.Unmarshal(v.Data, &msg); err != nil {
				log.Debugf("Unable to parse message due to invalid JSON. \nMessage:\n  %s\nError:\n  %s", v.Data, err)
			} else {
				log.Debugf("Broadcast %v", msg)
				hub.stream_broadcast <- msg
			}
		case redis.Subscription:
			log.Debugf("%s: %s %d\n", v.Channel, v.Kind, v.Count)
		case error:
			break
		}
	}
}
