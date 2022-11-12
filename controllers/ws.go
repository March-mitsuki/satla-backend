package controllers

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/March-mitsuki/satla-backend/utils/logger"

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
var allAutoCtxs autoCtxs = make(autoCtxs, 0)

func (s subscription) readPump() {
	c := s.conn
	var cUname string
	defer func() {
		err := allRoomUsers.delUser(s.room, cUname)
		if err != nil {
			// 如果发生错误则直接关闭连接并返回
			WsHub.unregister <- s
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
			logger.Err("ws", fmt.Sprintf("ws read msg err: %v /n", err))
			return
		}
		m := message{msg, s.room, s.conn}
		cmd := json.Get(msg, "head", "cmd").ToString()
		// switch根据发过来的cmd不同进行不同的处理
		switch cmd {
		case c2sCmdAddSubtitleUp:
			logger.Nomal("ws", "c2s: Cmd Add Subtitle Up")
			err := m.handleAddSubtitleUp()
			if err != nil {
				logger.Err("ws", fmt.Sprintf("add subtitle up err: %v \n", err))
				return
			}
			WsHub.broadcast <- m

		case c2sCmdAddSubtitleDown:
			logger.Nomal("ws", "c2s: Cmd Add Subtitle Down")
			err := m.handleAddSubtitleDown()
			if err != nil {
				logger.Err("ws", fmt.Sprintf("add subtitle down err: %v \n", err))
				return
			}
			WsHub.broadcast <- m

		case c2sCmdChangeUser:
			logger.Nomal("ws", "c2s: Cmd Change User")
			_cUname, addUserErr := m.handleAddUser()
			if addUserErr != nil {
				logger.Err("ws", fmt.Sprintf("add user err: %v \n", addUserErr))
				return
			}
			cUname = _cUname
			WsHub.broadcast <- m

		case c2sCmdGetRoomSubtitles:
			logger.Nomal("ws", "c2s: Cmd Get Room Subtitles")
			err := m.handleGetRoomSubtitles()
			if err != nil {
				logger.Err("ws", fmt.Sprintf("get all subtitles err %v \n", err))
				return
			}
			WsHub.castself <- m

		case c2sCmdChangeSubtitle:
			logger.Nomal("ws", "c2s: Cmd Change Subtitle")
			err := m.handleChangeSubtitle()
			if err != nil {
				logger.Err("ws", fmt.Sprintf("change subtitles err %v \n", err))
				return
			}
			WsHub.broadcast <- m

		case c2sCmdEditStart:
			logger.Nomal("ws", "c2s: Cmd Edit Start")
			err := m.handleEditStart()
			if err != nil {
				logger.Err("ws", fmt.Sprintf("edit start err %v \n", err))
				return
			}
			WsHub.broadcast <- m

		case c2sCmdEditEnd:
			logger.Nomal("ws", "c2s: Cmd Edit End")
			err := m.handleEditEnd()
			if err != nil {
				logger.Err("ws", fmt.Sprintf("edit end err %v \n", err))
				return
			}
			WsHub.broadcast <- m

		case c2sCmdAddTranslatedSub:
			logger.Nomal("ws", "c2s: Cmd Add Translated Sub")
			err := m.handleAddTranslatedSub()
			if err != nil {
				logger.Err("ws", fmt.Sprintf("add translated sub err %v \n", err))
				return
			}
			WsHub.broadcast <- m

		case c2sCmdDeleteSubtitle:
			logger.Nomal("ws", "c2s: Cmd Delete Subtitle")
			err := m.handleDeleteSubtitle()
			if err != nil {
				logger.Err("ws", fmt.Sprintf("delete subtitle err %v \n", err))
				return
			}
			WsHub.broadcast <- m

		case c2sCmdReorderSubFront:
			logger.Nomal("ws", "c2s: Cmd Reorder Sub Front")
			err := m.handleReorderSubFront()
			if err != nil {
				logger.Err("ws", fmt.Sprintf("reorder sub front err %v \n", err))
				return
			}
			WsHub.broadcast <- m

		case c2sCmdReorderSubBack:
			logger.Nomal("ws", "c2s: Cmd Reorder Sub Back")
			err := m.handleReorderSubBack()
			if err != nil {
				logger.Err("ws", fmt.Sprintf("reorder sub back err %v \n", err))
				return
			}
			WsHub.broadcast <- m

		case c2sCmdSendSubtitle:
			logger.Nomal("ws", "c2s: Cmd Send Subtitle")
			err := m.handleSendSubtitle()
			if err != nil {
				logger.Err("ws", fmt.Sprintf("send subtitle err %v \n", err))
				return
			}
			WsHub.broadcast <- m

		case c2sCmdSendSubtitleDirect:
			logger.Nomal("ws", "c2s: Cmd Send Subtitle Direct")
			err := m.handleSendSubtitleDirect()
			if err != nil {
				logger.Err("ws", fmt.Sprintf("send subtitle directly err %v \n", err))
				return
			}
			WsHub.broadcast <- m

		case c2sCmdChangeStyle:
			logger.Nomal("ws", "c2s: Cmd Change Style")
			err := m.handleChangeStyle()
			if err != nil {
				logger.Err("ws", fmt.Sprintf("change style err %v \n", err))
				return
			}
			WsHub.broadcast <- m

		case c2sCmdChangeBilingual:
			logger.Nomal("ws", "c2s: Cmd Change Bilingual")
			err := m.handleChangeBilingual()
			if err != nil {
				logger.Err("ws", fmt.Sprintf("change bilingual err %v \n", err))
				return
			}
			WsHub.broadcast <- m

		case c2sCmdChangeReversed:
			logger.Nomal("ws", "c2s: Cmd Change Reversed")
			err := m.handleChangeReversed()
			if err != nil {
				logger.Err("ws", fmt.Sprintf("change reversed err %v \n", err))
				return
			}
			WsHub.broadcast <- m

		case c2sCmdGetAutoLists:
			// 从这里往下是auto page
			logger.Nomal("ws", "c2s Cmd Get Auto Lists")
			err := m.handleGetRoomAutoLists()
			if err != nil {
				logger.Err("ws", fmt.Sprintf("get auto lists err %v \n", err))
				return
			}
			WsHub.castself <- m

		case c2sCmdAddAutoSub:
			logger.Nomal("ws", "c2s: Cmd Add Auto Sub")
			err := m.handleAddAutoSub()
			if err != nil {
				logger.Err("ws", fmt.Sprintf("add auto sub err %v \n", err))
				return
			}
			WsHub.broadcast <- m

		case c2sCmdPlayStart:
			logger.Nomal("ws", "c2s Cmd Play Start")

			autoCtx, endPlay := context.WithCancel(context.Background())
			listId := json.Get(msg, "body", "list_id").ToUint()
			ope := make(chan autoOpeData)
			ctxData := autoCtxData{autoCtx, endPlay, listId, ope}
			allAutoCtxs.addCtx(s.room, ctxData)

			err := m.handlePlayStart(autoCtx, ope)
			if err != nil {
				logger.Err("ws", fmt.Sprintf("auto play start err %v \n", err))
				return
			}

		case c2sCmdPlayEnd:
			logger.Nomal("ws", "c2s Cmd Play End")
			listId := json.Get(msg, "body", "list_id").ToUint()
			currentCtx, ctxErr := allAutoCtxs.getCurrentCtx(s.room, listId)
			if ctxErr != nil {
				logger.Err("ws", fmt.Sprintf("cmd play end getCurrentCtx %v \n", ctxErr))
				return
			}
			currentCtx.cancel()

		case c2sCmdPlayForward:
			logger.Nomal("ws", "c2s Cmd Play Forward")
			listId := json.Get(msg, "body", "list_id").ToUint()
			currentCtx, ctxErr := allAutoCtxs.getCurrentCtx(s.room, listId)
			if ctxErr != nil {
				logger.Err("ws", fmt.Sprintf("cmd play end getCurrentCtx %v \n", ctxErr))
				return
			}
			currentCtx.opeChan <- autoOpeData{
				opeType: foward,
			}

		case c2sCmdPlayForwardTwice:
			logger.Nomal("ws", "c2s Cmd Play Forward Twice")
			listId := json.Get(msg, "body", "list_id").ToUint()
			currentCtx, ctxErr := allAutoCtxs.getCurrentCtx(s.room, listId)
			if ctxErr != nil {
				logger.Err("ws", fmt.Sprintf("cmd play end getCurrentCtx %v \n", ctxErr))
				return
			}
			currentCtx.opeChan <- autoOpeData{
				opeType: fowardTwice,
			}

		case c2sCmdPlayRewind:
			logger.Nomal("ws", "c2s Cmd Play Rewind")
			listId := json.Get(msg, "body", "list_id").ToUint()
			currentCtx, ctxErr := allAutoCtxs.getCurrentCtx(s.room, listId)
			if ctxErr != nil {
				logger.Err("ws", fmt.Sprintf("cmd play end getCurrentCtx %v \n", ctxErr))
				return
			}
			currentCtx.opeChan <- autoOpeData{
				opeType: rewind,
			}

		case c2sCmdPlayRewindTwice:
			logger.Nomal("ws", "c2s Cmd Play Rewind Twice")
			listId := json.Get(msg, "body", "list_id").ToUint()
			currentCtx, ctxErr := allAutoCtxs.getCurrentCtx(s.room, listId)
			if ctxErr != nil {
				logger.Err("ws", fmt.Sprintf("cmd play end getCurrentCtx %v \n", ctxErr))
				return
			}
			currentCtx.opeChan <- autoOpeData{
				opeType: rewindTwice,
			}

		case c2sCmdPlayPause:
			logger.Nomal("ws", "c2s Cmd Play Pause")
			listId := json.Get(msg, "body", "list_id").ToUint()
			currentCtx, ctxErr := allAutoCtxs.getCurrentCtx(s.room, listId)
			if ctxErr != nil {
				logger.Err("ws", fmt.Sprintf("cmd play end getCurrentCtx %v \n", ctxErr))
				return
			}
			currentCtx.opeChan <- autoOpeData{
				opeType: pause,
			}

		case c2sCmdPlayRestart:
			logger.Nomal("ws", "c2s Cmd Play Restart")
			listId := json.Get(msg, "body", "list_id").ToUint()
			currentCtx, ctxErr := allAutoCtxs.getCurrentCtx(s.room, listId)
			if ctxErr != nil {
				logger.Err("ws", fmt.Sprintf("cmd play end getCurrentCtx %v \n", ctxErr))
				return
			}
			currentCtx.opeChan <- autoOpeData{
				opeType: restart,
			}

		case c2sCmdHeartBeat:
			logger.Info(
				"ws",
				"heart beat",
				fmt.Sprintf("room: %v, user: %v", s.room, cUname),
				fmt.Sprintf("now allAutoCtxs: \n==%v==\n", allAutoCtxs),
			)
			WsHub.castself <- m

		default:
			logger.Err("ws", fmt.Sprintf("\n --undefined cmd-- \n %+v \n", string(msg)))
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

func WsController(c *gin.Context, wsroom string) {
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
	sub := subscription{conn: conn, room: wsroom}
	WsHub.register <- sub

	go sub.writePump()
	sub.readPump()
}
