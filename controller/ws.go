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
		WsHub.unregister <- s
		c.ws.Close()
		fmt.Println("ws close by read pump")
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
			WsHub.broadcast <- m
		case "addUser":
			WsHub.castother <- m
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
		fmt.Println("ws close by write punp")
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
	fmt.Println("ws controller called")
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		ws.Close()
		fmt.Println("ws close by ws controller")
	}()

	conn := &connection{send: make(chan []byte, 256), ws: ws}
	sub := subscription{conn: conn, room: roomid}
	WsHub.register <- sub

	go sub.writePump()
	sub.readPump()
}
