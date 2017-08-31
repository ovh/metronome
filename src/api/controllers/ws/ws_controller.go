package wsCtrl

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/websocket"
	"github.com/ovh/metronome/src/api/core/ws"

	"github.com/ovh/metronome/src/api/core/io/out"
	authSrv "github.com/ovh/metronome/src/api/services/auth"
	"github.com/ovh/metronome/src/metronome/redis"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// Join handle ws connections.
func Join(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error(err)
		return
	}
	client := ws.NewClient(conn)
	defer client.Close()

	// wait for auth token
	msg, ok := <-client.Messages()
	if !ok {
		return
	}

	token := authSrv.GetToken(msg)
	if token == nil {
		out.Unauthorized(w)
		return
	}

	pubsub, err := redis.DB().Subscribe(authSrv.UserID(token))
	if err != nil {
		log.Error(err)
		out.BadGateway(w)
		return
	}
	defer pubsub.Close()

	in := make(chan string)
	kill := make(chan struct{})

	go func() {
		for {
			msg, err := pubsub.ReceiveMessage()
			if err != nil {
				kill <- struct{}{}
				return
			}

			in <- msg.Payload
		}
	}()

	for {
		select {
		case _, ok := <-client.Messages():
			if !ok { // shuting down
				out.Success(w)
				return
			}

		case msg := <-in:
			client.Send(msg)

		case <-kill:
			out.BadGateway(w)
			return
		}
	}
}
