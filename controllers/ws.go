package controllers

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

var allRoomUsers roomUsers = make(roomUsers, 0)

func (s subscription) readPump() {
	c := s.conn
	var cUname string
	defer func() {
		WsHub.unregister <- s
		allRoomUsers.delUser(s.room, cUname)
		c.ws.Close()
		fmt.Println("ws close by read pump")
	}()

	for {
		_, msg, err := c.ws.ReadMessage()
		if err != nil {
			return
		}
		m := message{msg, s.room, s.conn}
		cmd := json.Get(msg, "head", "cmd").ToString()
		switch cmd {
		case c2sCmdAddSubtitle:
			var wsData c2sSubtitle
			json.Unmarshal(msg, &wsData)
			fmt.Printf("\n --parse add subtitle-- \n %+v \n", wsData)
			WsHub.broadcast <- m
		case c2sCmdAddUser:
			_cUname, addUserErr := m.handleAddUser()
			if addUserErr != nil {
				fmt.Printf("add user err: %v \n", addUserErr)
				return
			}
			cUname = _cUname
			WsHub.broadcast <- m
		default:
			fmt.Printf("\n --undefined cmd-- \n %+v \n", string(msg))
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
			fmt.Println("<-c.send err")
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
