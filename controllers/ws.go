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
		allRoomUsers.delUser(s.room, cUname)
		_closeMsgData := s2cChangeUser{
			Head: struct {
				Cmd s2cCmds "json:\"cmd\""
			}{
				Cmd: s2cCmdChangeUser,
			},
			Body: struct {
				Users []string "json:\"users\""
			}{
				Users: allRoomUsers[s.room],
			},
		}
		closeMsgData, marshalErr := json.Marshal(&_closeMsgData)
		if marshalErr != nil {
			fmt.Println("unregister change user err")
			WsHub.unregister <- s
			c.ws.Close()
		}
		closeMsg := message{closeMsgData, s.room, s.conn}
		WsHub.broadcast <- closeMsg
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
		cmd := json.Get(msg, "head", "cmd").ToString()
		// switch根据发过来的cmd不同进行不同的处理
		switch cmd {
		case c2sCmdAddSubtitleUp:
			err := m.handleAddSubtitleUp()
			if err != nil {
				fmt.Printf("add subtitle up err: %v \n", err)
				return
			}
			WsHub.broadcast <- m
		case c2sCmdAddSubtitleDown:
			err := m.handleAddSubtitleDown()
			if err != nil {
				fmt.Printf("add subtitle down err: %v \n", err)
				return
			}
			WsHub.broadcast <- m
		case c2sCmdChangeUser:
			_cUname, addUserErr := m.handleAddUser()
			if addUserErr != nil {
				fmt.Printf("add user err: %v \n", addUserErr)
				return
			}
			cUname = _cUname
			WsHub.broadcast <- m
		case c2sCmdGetRoomSubtitles:
			err := m.handleGetRoomSubtitles()
			if err != nil {
				fmt.Printf("get all subtitles err %v \n", err)
				return
			}
			WsHub.castself <- m
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
		// writePump将readPump传过来的message原封不动写给client
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
