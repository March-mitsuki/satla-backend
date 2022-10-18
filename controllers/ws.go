package controllers

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == os.Getenv("CORS_ORIGIN") {
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
		err := allRoomUsers.delUser(s.room, cUname)
		if err != nil {
			// 如果发生错误则直接关闭连接并返回
			c.ws.Close()
			fmt.Println("ws close by read pump err close")
			return
		}
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
		return
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
			fmt.Println("--- c2s: Cmd Add Subtitle Up ---")

			err := m.handleAddSubtitleUp()
			if err != nil {
				fmt.Printf("add subtitle up err: %v \n", err)
				return
			}
			WsHub.broadcast <- m
		case c2sCmdAddSubtitleDown:
			fmt.Println("--- c2s: Cmd Add Subtitle Down ---")

			err := m.handleAddSubtitleDown()
			if err != nil {
				fmt.Printf("add subtitle down err: %v \n", err)
				return
			}
			WsHub.broadcast <- m
		case c2sCmdChangeUser:
			fmt.Println("--- c2s: Cmd Change User ---")

			_cUname, addUserErr := m.handleAddUser()
			if addUserErr != nil {
				fmt.Printf("add user err: %v \n", addUserErr)
				return
			}
			cUname = _cUname
			WsHub.broadcast <- m
		case c2sCmdGetRoomSubtitles:
			fmt.Println("--- c2s: Cmd Get Room Subtitles ---")

			err := m.handleGetRoomSubtitles()
			if err != nil {
				fmt.Printf("get all subtitles err %v \n", err)
				return
			}
			WsHub.castself <- m
		case c2sCmdChangeSubtitle:
			fmt.Println("--- c2s: Cmd Change Subtitle ---")

			err := m.handleChangeSubtitle()
			if err != nil {
				fmt.Printf("change subtitles err %v \n", err)
				return
			}
			WsHub.broadcast <- m
		case c2sCmdEditStart:
			fmt.Println("--- c2s: Cmd Edit Start ---")

			err := m.handleEditStart()
			if err != nil {
				fmt.Printf("edit start err %v \n", err)
				return
			}
			WsHub.broadcast <- m
		case c2sCmdEditEnd:
			fmt.Println("--- c2s: Cmd Edit End ---")

			err := m.handleEditEnd()
			if err != nil {
				fmt.Printf("edit end err %v \n", err)
				return
			}
			WsHub.broadcast <- m
		case c2sCmdAddTranslatedSub:
			fmt.Println("--- c2s: Cmd Add Translated Sub ---")

			err := m.handleAddTranslatedSub()
			if err != nil {
				fmt.Printf("add translated sub err %v \n", err)
				return
			}
			WsHub.broadcast <- m
		case c2sCmdDeleteSubtitle:
			fmt.Println("--- c2s: Cmd Delete Subtitle ---")

			err := m.handleDeleteSubtitle()
			if err != nil {
				fmt.Printf("delete subtitle err %v \n", err)
				return
			}
			WsHub.broadcast <- m
		case c2sCmdReorderSubFront:
			fmt.Println("--- c2s: Cmd Reorder Sub Front ---")

			err := m.handleReorderSubFront()
			if err != nil {
				fmt.Printf("reorder sub front err %v \n", err)
				return
			}
			WsHub.broadcast <- m
		case c2sCmdReorderSubBack:
			fmt.Println("--- c2s: Cmd Reorder Sub Back ---")

			err := m.handleReorderSubBack()
			if err != nil {
				fmt.Printf("reorder sub back err %v \n", err)
				return
			}
			WsHub.broadcast <- m
		case c2sCmdSendSubtitle:
			fmt.Println("--- c2s: Cmd Send Subtitle ---")

			err := m.handleSendSubtitle()
			if err != nil {
				fmt.Printf("send subtitle err %v \n", err)
				return
			}
			WsHub.broadcast <- m
		case c2sCmdSendSubtitleDirect:
			fmt.Println("--- c2s: Cmd Send Subtitle Direct ---")

			err := m.handleSendSubtitleDirect()
			if err != nil {
				fmt.Printf("send subtitle directly err %v \n", err)
				return
			}
			WsHub.broadcast <- m
		case c2sCmdChangeStyle:
			fmt.Println("--- c2s: Cmd Change Style ---")

			err := m.handleChangeStyle()
			if err != nil {
				fmt.Printf("change style err %v \n", err)
				return
			}
			WsHub.broadcast <- m
		case c2sCmdChangeBilingual:
			fmt.Println("--- c2s: Cmd Change Bilingual ---")

			err := m.handleChangeBilingual()
			if err != nil {
				fmt.Printf("change bilingual err %v \n", err)
				return
			}
			WsHub.broadcast <- m
		case c2sCmdChangeReversed:
			fmt.Println("--- c2s: Cmd Change Reversed ---")

			err := m.handleChangeReversed()
			if err != nil {
				fmt.Printf("change reversed err %v \n", err)
				return
			}
			WsHub.broadcast <- m
		case c2sCmdHeartBeat:
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
