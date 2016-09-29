package webapp

import (
	"log"
	"net/http"
	"time"

	"github.com/eleme/banshee/models"
	"github.com/gorilla/websocket"
)

var (
	writeWait = 10 * time.Second

	pongWait = 60 * time.Second

	pingPeriod = (pongWait * 9) / 10

	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

type WsClient struct {
	conn *websocket.Conn
	send chan []byte
}

func (s *WsClient) readPump() {
	defer s.close()
	s.conn.SetReadDeadline(time.Now().Add(pongWait))
	s.conn.SetPongHandler(func(string) error { s.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, _, err := s.conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (s *WsClient) writePump() {
	events := make(chan models.EventWrapper, 1024)
	handler := func(ew *models.EventWrapper) {
		events <- ew
	}
	bus.Subscribe("alert", handler)
	defer bus.UnSubscribe("alert", handler)
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		s.close()
	}()
	for {
		select {
		case message, ok := <-s.send:
			if !ok {
				s.write(websocket.CloseMessage, []byte{})
				return
			}
			s.conn.SetWriteDeadline(time.Now().Add(writeWait))
			w, err := s.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(s.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-s.send)
			}
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			if err := s.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}

	}
}

func (s *WsClient) write(mt int, payload []byte) error {
	s.conn.SetWriteDeadline(time.Now().Add(writeWait))
	return s.conn.WriteMessage(mt, payload)
}

func (s *WsClient) close() {
	s.conn.Close()
}
func serverWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			log.Println(err)
		}
		return
	}
	server := &WsClient{conn: conn, send: make(chan []byte, 256)}
	go server.writePump()
	server.readPump()
}
