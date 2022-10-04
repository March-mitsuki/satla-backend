package controller

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "http://192.168.64.3:8080" || origin == "http://localhost:3131" {
			return true
		} else {
			return false
		}
	},
}

func (s subscription) readPump() {
	c := s.conn
	defer func() {
		h.unregister <- s
		c.ws.Close()
	}()

	for {
		_, msg, err := c.ws.ReadMessage()
		if err != nil {
			return
		}
		m := message{msg, s.room, s.conn}
		var unmarshal subtitleData
		json.Unmarshal(msg, &unmarshal)
		fmt.Println("-----unmarshal------")
		fmt.Println(unmarshal.Body)
		cmd := unmarshal.Head.Cmd
		switch cmd {
		case "addSubtitle":
			h.broadcast <- m
		case "addUser":
			h.castother <- m
		}
	}
}

func (c *connection) write(mt int, payload []byte) error {
	return c.ws.WriteMessage(mt, payload)
}

func (s subscription) writePump() {
	c := s.conn
	defer func() {
		c.ws.Close()
	}()

	for {
		msg, ok := <-c.send
		if !ok {
			c.write(websocket.CloseMessage, []byte("<-c.send err"))
			return
		}
		fmt.Println("send once")
		if err := c.write(websocket.TextMessage, msg); err != nil {
			return
		}
	}
}

func WsController(c *gin.Context, roomid string) {
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer ws.Close()
	defer fmt.Println("ws closed")

	conn := &connection{send: make(chan []byte, 256), ws: ws}
	sub := subscription{conn: conn, room: roomid}
	h.register <- sub

	go sub.writePump()
	sub.readPump()
}
