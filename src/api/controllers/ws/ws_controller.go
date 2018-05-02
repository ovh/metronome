package wsctrl

import (
	"errors"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/ovh/metronome/src/api/core/ws"
	"github.com/ovh/metronome/src/api/factories"
	log "github.com/sirupsen/logrus"

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
		log.WithError(err).Error("Could not upgrade the http request to websocket")
		out.JSON(w, http.StatusInternalServerError, factories.Error(err))
		return
	}
	client := ws.NewClient(conn)
	defer client.Close()

	// wait for auth token
	msg, ok := <-client.Messages()
	if !ok {
		return
	}

	token, err := authSrv.GetToken(msg)
	if err != nil {
		out.JSON(w, http.StatusInternalServerError, factories.Error(err))
		return
	}

	if token == nil {
		out.JSON(w, http.StatusUnauthorized, factories.Error(errors.New("Unauthorized")))
		return
	}

	pubsub, err := redis.DB().Subscribe(authSrv.UserID(token))
	if err != nil {
		log.WithError(err).Error("Could not subscribe to redis")
		out.JSON(w, http.StatusInternalServerError, factories.Error(err))
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
				w.WriteHeader(http.StatusOK)
				return
			}

		case msg := <-in:
			client.Send(msg)

		case <-kill:
			out.JSON(w, http.StatusBadGateway, factories.Error(errors.New("Bad gateway")))
			return
		}
	}
}
