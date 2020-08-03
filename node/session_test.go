package node

import (
	"sync"
	"testing"
)

func TestSendRaceConditions(t *testing.T) {
	var wg sync.WaitGroup

	for i := 1; i <= 10; i++ {
		session := NewMockSession("123")
		// small buffer channel
		session.send = make(chan sentFrame, 1)

		wg.Add(2)
		go func() {
			go func() {
				session.Send([]byte("hi!"), false)
				wg.Done()
			}()

			go func() {
				session.Send([]byte("bye"), false)
				wg.Done()
			}()
		}()

		wg.Add(2)
		go func() {
			go func() {
				session.Send([]byte("bye"), false)
				wg.Done()
			}()

			go func() {
				session.Send([]byte("why"), false)
				wg.Done()
			}()
		}()
	}

	wg.Wait()
}
