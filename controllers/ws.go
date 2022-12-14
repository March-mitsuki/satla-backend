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
		// 通知房间内其他人更改用户列表
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
		// 所有处理(包括发送)都在handler内完结, 此处只负责log意外
		// handler内处理意外要一条一条写, 这里直接抓就行
		switch cmd {
		case c2sCmdAddSubtitleUp:
			logger.Nomal("ws", "c2s: Cmd Add Subtitle Up")
			err := m.handleAddSubtitleUp()
			if err != nil {
				logger.Err("ws", fmt.Sprintf("add subtitle up err: %v \n", err))
				return
			}

		case c2sCmdAddSubtitleDown:
			logger.Nomal("ws", "c2s: Cmd Add Subtitle Down")
			err := m.handleAddSubtitleDown()
			if err != nil {
				logger.Err("ws", fmt.Sprintf("add subtitle down err: %v \n", err))
				return
			}

		case c2sCmdChangeUser:
			logger.Nomal("ws", "c2s: Cmd Change User")
			_cUname, addUserErr := m.handleAddUser()
			if addUserErr != nil {
				logger.Err("ws", fmt.Sprintf("add user err: %v \n", addUserErr))
				return
			}
			cUname = _cUname

		case c2sCmdGetRoomSubtitles:
			logger.Nomal("ws", "c2s: Cmd Get Room Subtitles")
			err := m.handleGetRoomSubtitles()
			if err != nil {
				logger.Err("ws", fmt.Sprintf("get all subtitles err %v \n", err))
				return
			}

		case c2sCmdChangeSubtitle:
			logger.Nomal("ws", "c2s: Cmd Change Subtitle")
			err := m.handleChangeSubtitle()
			if err != nil {
				logger.Err("ws", fmt.Sprintf("change subtitles err %v \n", err))
				return
			}

		case c2sCmdEditStart:
			logger.Nomal("ws", "c2s: Cmd Edit Start")
			err := m.handleEditStart()
			if err != nil {
				logger.Err("ws", fmt.Sprintf("edit start err %v \n", err))
				return
			}

		case c2sCmdEditEnd:
			logger.Nomal("ws", "c2s: Cmd Edit End")
			err := m.handleEditEnd()
			if err != nil {
				logger.Err("ws", fmt.Sprintf("edit end err %v \n", err))
				return
			}

		case c2sCmdAddTranslatedSub:
			logger.Nomal("ws", "c2s: Cmd Add Translated Sub")
			err := m.handleAddTranslatedSub()
			if err != nil {
				logger.Err("ws", fmt.Sprintf("add translated sub err %v \n", err))
				return
			}

		case c2sCmdDeleteSubtitle:
			logger.Nomal("ws", "c2s: Cmd Delete Subtitle")
			err := m.handleDeleteSubtitle()
			if err != nil {
				logger.Err("ws", fmt.Sprintf("delete subtitle err %v \n", err))
				return
			}

		case c2sCmdReorderSubFront:
			logger.Nomal("ws", "c2s: Cmd Reorder Sub Front")
			err := m.handleReorderSubFront()
			if err != nil {
				logger.Err("ws", fmt.Sprintf("reorder sub front err %v \n", err))
				return
			}

		case c2sCmdReorderSubBack:
			logger.Nomal("ws", "c2s: Cmd Reorder Sub Back")
			err := m.handleReorderSubBack()
			if err != nil {
				logger.Err("ws", fmt.Sprintf("reorder sub back err %v \n", err))
				return
			}

		case c2sCmdSendSubtitle:
			logger.Nomal("ws", "c2s: Cmd Send Subtitle")
			err := m.handleSendSubtitle()
			if err != nil {
				logger.Err("ws", fmt.Sprintf("send subtitle err %v \n", err))
				return
			}

		case c2sCmdSendSubtitleDirect:
			logger.Nomal("ws", "c2s: Cmd Send Subtitle Direct")
			err := m.handleSendSubtitleDirect()
			if err != nil {
				logger.Err("ws", fmt.Sprintf("send subtitle directly err %v \n", err))
				return
			}

		case c2sCmdChangeStyle:
			logger.Nomal("ws", "c2s: Cmd Change Style")
			err := m.handleChangeStyle()
			if err != nil {
				logger.Err("ws", fmt.Sprintf("change style err %v \n", err))
				return
			}

		case c2sCmdGetNowRoomStyle:
			logger.Nomal("ws", "c2s Cmd Get Now Room Style")
			err := m.handleGetNowRoomStyle()
			if err != nil {
				logger.Err("ws", fmt.Sprintf("get now room style err %v \n", err))
				return
			}

		case c2sCmdGetNowRoomSub:
			logger.Nomal("ws", "c2s Cmd Get Now Room Sub")
			err := m.handleGetNowRoomSub()
			if err != nil {
				logger.Err("ws", fmt.Sprintf("get now room sub err %v \n", err))
				return
			}

		case c2sCmdGetAutoLists:
			//
			// 从这里往下是auto page
			//
			logger.Nomal("ws", "c2s Cmd Get Auto Lists")
			err := m.handleGetRoomAutoLists()
			if err != nil {
				logger.Err("ws", fmt.Sprintf("get auto lists err %v \n", err))
				return
			}

		case c2sCmdAddAutoSub:
			logger.Nomal("ws", "c2s: Cmd Add Auto Sub")
			err := m.handleAddAutoSub()
			if err != nil {
				logger.Err("ws", fmt.Sprintf("add auto sub err %v \n", err))
				return
			}

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
			// end 会找到当前播放的ctx并发出done信号
			// pause与手动挡的停止也使用这个借口
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

		case c2sCmdPlaySendBlank:
			logger.Nomal("ws", "c2s Cmd Play Send Space")
			broadcastSendBlank(&m)

		case c2sCmdAutoToManual:
			logger.Nomal("ws", "c2s Cmd Auto To Manual")
			listId := json.Get(msg, "body", "list_id").ToUint()
			currentCtx, ctxErr := allAutoCtxs.getCurrentCtx(s.room, listId)
			if ctxErr != nil {
				logger.Err("ws", fmt.Sprintf("cmd play end getCurrentCtx %v \n", ctxErr))
				return
			}
			currentCtx.opeChan <- autoOpeData{
				opeType: toManual,
			}

		case c2sCmdManualToAuto:
			logger.Nomal("ws", "c2s Cmd Manual To Auto")
			listId := json.Get(msg, "body", "list_id").ToUint()
			currentCtx, ctxErr := allAutoCtxs.getCurrentCtx(s.room, listId)
			if ctxErr != nil {
				logger.Err("ws", fmt.Sprintf("cmd play end getCurrentCtx %v \n", ctxErr))
				return
			}
			currentCtx.opeChan <- autoOpeData{
				opeType: toAuto,
			}

		case c2sCmdDeleteAutoSub:
			logger.Nomal("ws", "c2s Cmd Delete Auto Sub")
			err := m.handleDeleteAutoSub()
			if err != nil {
				logger.Err("ws", fmt.Sprintf("delete auto sub: %v \n", err))
				return
			}

		case c2sCmdGetAutoPlayStat:
			logger.Nomal("ws", "c2s Cmd Get Auto Play Stat")
			rdbCtx := context.Background()
			err := m.handleGetAutoPlayStat(rdbCtx)
			if err != nil {
				logger.Err("ws", fmt.Sprintf("get auto play stat err: %v \n", err))
				return
			}

		case c2sCmdRecoverPlayStat:
			// 初始化房间会清除当前房间内的播放(如果正在播放)
			// 并删除储存在redis中的房间stat, 以及全部的ctx
			// 并更改该房间内所有List为未播放
			logger.Nomal("ws", "c2s Cmd Recover Play Stat")
			allAutoCtxs.delRoom(m.room)

			rdbCtx := context.Background()
			err := m.handleRecoverPlayStat(rdbCtx)
			if err != nil {
				logger.Err("ws", fmt.Sprintf("get auto play stat err: %v \n", err))
				return
			}

		case c2sCmdChangeAutoMemo:
			logger.Nomal("ws", "c2s Cmd Change Auto Memo")
			err := m.handleChangeAutoMemo()
			if err != nil {
				logger.Err("ws", fmt.Sprintf("change auto memo err: %v \n", err))
				return
			}

		case c2sCmdBatchAddSubs:
			logger.Nomal("ws", "c2s Cmd Batch Add Subs")
			err := m.handleBatchAddSubs()
			if err != nil {
				logger.Err("ws", fmt.Sprintf("batch add subs err: %v", err))
				return
			}

		case c2sCmdHeartBeat:
			if os.Getenv("GIN_MODE") != "release" {
				logger.Info(
					"ws",
					"heart beat",
					fmt.Sprintf("room: %v, user: %v", s.room, cUname),
					fmt.Sprintf("now allAutoCtxs: \n==%v==\n", allAutoCtxs),
				)
			}
			err := m.handleHeartBeat()
			if err != nil {
				logger.Err("ws", fmt.Sprintf("heart beat check err: %v", err))
				return
			}

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

	conn := &connection{
		send: make(chan []byte, 1024),
		ws:   ws,
	}
	sub := subscription{conn: conn, room: wsroom}
	WsHub.register <- sub

	go sub.writePump()
	sub.readPump()
}
