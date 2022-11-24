package controllers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/March-mitsuki/satla-backend/controllers/db"
	"github.com/March-mitsuki/satla-backend/model"
	"github.com/March-mitsuki/satla-backend/utils/logger"
	"github.com/March-mitsuki/satla-backend/utils/stat"
	"github.com/go-redis/redis/v9"
)

func (m *message) handleAddUser() (string, error) {
	// 除了更改m中data的内容,并返回error之外
	// 还会额外返回一个供删除时使用的username(string)
	var wsData c2sChangeUser
	unmarshalErr := json.Unmarshal(m.data, &wsData)
	if unmarshalErr != nil {
		return "", unmarshalErr
	}
	// fmt.Printf("\n --parse add user-- \n %+v \n", wsData)
	allRoomUsers.addUser(m.room, wsData.Body.Uname)
	_data := s2cChangeUser{
		Head: struct {
			Cmd s2cCmds "json:\"cmd\""
		}{
			Cmd: s2cCmdChangeUser,
		},
		Body: struct {
			Users []string "json:\"users\""
		}{
			Users: allRoomUsers[m.room],
		},
	}
	data, marshalErr := json.Marshal(&_data)
	if marshalErr != nil {
		return "", marshalErr
	}
	m.data = data

	WsHub.broadcast <- *m

	return wsData.Body.Uname, nil
}

func (m *message) handleGetRoomSubtitles() error {
	var wsData c2sGetRoomSubtitles
	unmarshalErr := json.Unmarshal(m.data, &wsData)
	if unmarshalErr != nil {
		return unmarshalErr
	}
	subtitles, order, dbGetErr := db.GetRoomSubtitles(wsData.Body.RoomId)
	if dbGetErr != nil {
		return dbGetErr
	}
	_data := s2cGetRoomSubtitles{
		Head: struct {
			Cmd s2cCmds "json:\"cmd\""
		}{
			Cmd: s2cCmdGetRoomSubtitles,
		},
		Body: struct {
			Subtitles []model.Subtitle "json:\"subtitles\""
			Order     string           "json:\"order\""
		}{
			Subtitles: subtitles,
			Order:     order,
		},
	}
	data, marshalErr := json.Marshal(&_data)
	if marshalErr != nil {
		return marshalErr
	}
	m.data = data

	WsHub.castself <- *m

	return nil
}

func (m *message) handleAddSubtitleUp() error {
	var wsData c2sAddSubtitle
	unmarshalErr := json.Unmarshal(m.data, &wsData)
	if unmarshalErr != nil {
		return unmarshalErr
	}
	arg := db.ArgAddSubtitle{
		PreSubtitleId: wsData.Body.PreSubtitleId,
		RoomId:        wsData.Body.RoomId,
		CheckedBy:     wsData.Body.CheckedBy,
	}
	newSubId, err := db.CreateSubtitleUp(arg)
	if err != nil {
		return err
	}
	_data := s2cAddSubtitle{
		Head: struct {
			Cmd s2cCmds "json:\"cmd\""
		}{
			Cmd: s2cCmdAddSubtitleUp,
		},
		Body: struct {
			RoomId         uint   "json:\"room_id\""
			NewSubtitleId  uint   "json:\"new_subtitle_id\""
			PreSubtitleIdx uint   "json:\"pre_subtitle_idx\""
			CheckedBy      string "json:\"checked_by\""
		}{
			RoomId:         wsData.Body.RoomId,
			PreSubtitleIdx: wsData.Body.PreSubtitleIdx,
			NewSubtitleId:  newSubId,
			CheckedBy:      wsData.Body.CheckedBy,
		},
	}
	data, marshalErr := json.Marshal(&_data)
	if marshalErr != nil {
		return marshalErr
	}
	m.data = data

	WsHub.broadcast <- *m

	return nil
}

func (m *message) handleAddSubtitleDown() error {
	var wsData c2sAddSubtitle
	unmarshalErr := json.Unmarshal(m.data, &wsData)
	if unmarshalErr != nil {
		return unmarshalErr
	}
	arg := db.ArgAddSubtitle{
		PreSubtitleId: wsData.Body.PreSubtitleId,
		RoomId:        wsData.Body.RoomId,
		CheckedBy:     wsData.Body.CheckedBy,
	}
	newSubId, err := db.CreateSubtitleDown(arg)
	if err != nil {
		return err
	}
	_data := s2cAddSubtitle{
		Head: struct {
			Cmd s2cCmds "json:\"cmd\""
		}{
			Cmd: s2cCmdAddSubtitleDown,
		},
		Body: struct {
			RoomId         uint   "json:\"room_id\""
			NewSubtitleId  uint   "json:\"new_subtitle_id\""
			PreSubtitleIdx uint   "json:\"pre_subtitle_idx\""
			CheckedBy      string "json:\"checked_by\""
		}{
			RoomId:         wsData.Body.RoomId,
			PreSubtitleIdx: wsData.Body.PreSubtitleIdx,
			NewSubtitleId:  newSubId,
			CheckedBy:      wsData.Body.CheckedBy,
		},
	}
	data, marshalErr := json.Marshal(&_data)
	if marshalErr != nil {
		return marshalErr
	}
	m.data = data

	WsHub.broadcast <- *m

	return nil
}

func (m *message) handleChangeSubtitle() error {
	var wsData c2sChangeSubtitle
	unmarshalErr := json.Unmarshal(m.data, &wsData)
	if unmarshalErr != nil {
		return unmarshalErr
	}
	arg := db.ArgChangeSubtitle{
		ID:        wsData.Body.ID,
		CheckedBy: wsData.Body.CheckedBy,
		Subtitle:  wsData.Body.Subtitle.Subtitle,
		Origin:    wsData.Body.Origin,
	}
	var _data s2cChangeSubtitle
	err := db.ChangeSubtitle(arg)
	if err != nil {
		// 若未能正确修改字幕则回复一个消息告知未能修改
		_data = s2cChangeSubtitle{
			Head: struct {
				Cmd s2cCmds "json:\"cmd\""
			}{
				Cmd: s2cCmdChangeSubtitle,
			},
			Body: struct {
				Status   bool           "json:\"status\""
				Subtitle model.Subtitle "json:\"subtitle\""
			}{
				Status:   false,
				Subtitle: wsData.Body.Subtitle,
			},
		}
	} else {
		_data = s2cChangeSubtitle{
			Head: struct {
				Cmd s2cCmds "json:\"cmd\""
			}{
				Cmd: s2cCmdChangeSubtitle,
			},
			Body: struct {
				Status   bool           "json:\"status\""
				Subtitle model.Subtitle "json:\"subtitle\""
			}{
				Status:   true,
				Subtitle: wsData.Body.Subtitle,
			},
		}
	}
	data, marshalErr := json.Marshal(&_data)
	if marshalErr != nil {
		return marshalErr
	}
	m.data = data

	WsHub.broadcast <- *m

	return nil
}

func (m *message) handleEditStart() error {
	var wsData c2sEditChange
	unmarshalErr := json.Unmarshal(m.data, &wsData)
	if unmarshalErr != nil {
		return unmarshalErr
	}
	_data := s2cEditChange{
		Head: struct {
			Cmd s2cCmds "json:\"cmd\""
		}{
			Cmd: s2cCmdEditStart,
		},
		Body: struct {
			Uname      string "json:\"uname\""
			SubtitleId uint   "json:\"subtitle_id\""
		}{
			Uname:      wsData.Body.Uname,
			SubtitleId: wsData.Body.SubtitleId,
		},
	}
	data, marshalErr := json.Marshal(&_data)
	if marshalErr != nil {
		return marshalErr
	}
	m.data = data

	WsHub.broadcast <- *m

	return nil
}

func (m *message) handleEditEnd() error {
	var wsData c2sEditChange
	unmarshalErr := json.Unmarshal(m.data, &wsData)
	if unmarshalErr != nil {
		return unmarshalErr
	}
	_data := s2cEditChange{
		Head: struct {
			Cmd s2cCmds "json:\"cmd\""
		}{
			Cmd: s2cCmdEditEnd,
		},
		Body: struct {
			Uname      string "json:\"uname\""
			SubtitleId uint   "json:\"subtitle_id\""
		}{
			Uname:      wsData.Body.Uname,
			SubtitleId: wsData.Body.SubtitleId,
		},
	}
	data, marshalErr := json.Marshal(&_data)
	if marshalErr != nil {
		return marshalErr
	}
	m.data = data

	WsHub.broadcast <- *m

	return nil
}

func (m *message) handleAddTranslatedSub() error {
	var wsData c2sAddTranslatedSub
	unmarshalErr := json.Unmarshal(m.data, &wsData)
	if unmarshalErr != nil {
		return unmarshalErr
	}
	// 这里收到的subtitle的id为0, 需要额外操作
	newSub, dbErr := db.CreateTranslatedSub(
		wsData.Body.NewSubtitle,
	)
	if dbErr != nil {
		return dbErr
	}
	_data := s2cAddTranslatedSub{
		Head: struct {
			Cmd s2cCmds "json:\"cmd\""
		}{
			Cmd: s2cCmdAddTranslatedSub,
		},
		Body: struct {
			NewSubtitle model.Subtitle "json:\"new_subtitle\""
		}{
			NewSubtitle: newSub,
		},
	}
	data, marshalErr := json.Marshal(&_data)
	if marshalErr != nil {
		return marshalErr
	}
	m.data = data

	WsHub.broadcast <- *m

	return nil
}

func (m *message) handleDeleteSubtitle() error {
	var wsData c2sDeleteSubtitle
	unmarshalErr := json.Unmarshal(m.data, &wsData)
	if unmarshalErr != nil {
		return unmarshalErr
	}
	var _data s2cDeleteSubtitle
	err := db.DeleteSubtitle(wsData.Body.Subtitle)
	if err != nil {
		_data = s2cDeleteSubtitle{
			Head: struct {
				Cmd s2cCmds "json:\"cmd\""
			}{
				Cmd: s2cCmdDeleteSubtitle,
			},
			Body: struct {
				Status     bool "json:\"status\""
				SubtitleId uint "json:\"subtitle_id\""
			}{
				Status:     false,
				SubtitleId: wsData.Body.Subtitle.ID,
			},
		}
	} else {
		_data = s2cDeleteSubtitle{
			Head: struct {
				Cmd s2cCmds "json:\"cmd\""
			}{
				Cmd: s2cCmdDeleteSubtitle,
			},
			Body: struct {
				Status     bool "json:\"status\""
				SubtitleId uint "json:\"subtitle_id\""
			}{
				Status:     true,
				SubtitleId: wsData.Body.Subtitle.ID,
			},
		}
	}
	data, marshalErr := json.Marshal(&_data)
	if marshalErr != nil {
		return marshalErr
	}
	m.data = data

	WsHub.broadcast <- *m

	return nil
}

func (m *message) handleReorderSubFront() error {
	var wsData c2sReorderSub
	unmarshalErr := json.Unmarshal(m.data, &wsData)
	if unmarshalErr != nil {
		return unmarshalErr
	}
	var _data s2cReorderSub
	err := db.ReorderSubtitle(
		wsData.Body.RoomId,
		wsData.Body.DragId,
		wsData.Body.DropId,
	)
	if err != nil {
		_data = s2cReorderSub{
			Head: struct {
				Cmd s2cCmds "json:\"cmd\""
			}{
				Cmd: s2cCmdReorderSubFront,
			},
			Body: struct {
				OperationUser string "json:\"operation_user\""
				Status        bool   "json:\"status\""
				DragId        uint   "json:\"drag_id\""
				DropId        uint   "json:\"drop_id\""
			}{
				OperationUser: wsData.Body.OperationUser,
				Status:        true,
				DragId:        wsData.Body.DragId,
				DropId:        wsData.Body.DropId,
			},
		}
	} else {
		_data = s2cReorderSub{
			Head: struct {
				Cmd s2cCmds "json:\"cmd\""
			}{
				Cmd: s2cCmdReorderSubFront,
			},
			Body: struct {
				OperationUser string "json:\"operation_user\""
				Status        bool   "json:\"status\""
				DragId        uint   "json:\"drag_id\""
				DropId        uint   "json:\"drop_id\""
			}{
				OperationUser: wsData.Body.OperationUser,
				Status:        true,
				DragId:        wsData.Body.DragId,
				DropId:        wsData.Body.DropId,
			},
		}
	}

	data, marshalErr := json.Marshal(&_data)
	if marshalErr != nil {
		return marshalErr
	}
	m.data = data

	WsHub.broadcast <- *m

	return nil
}

func (m *message) handleReorderSubBack() error {
	var wsData c2sReorderSub
	unmarshalErr := json.Unmarshal(m.data, &wsData)
	if unmarshalErr != nil {
		return unmarshalErr
	}
	var _data s2cReorderSub
	err := db.ReorderSubtitle(
		wsData.Body.RoomId,
		wsData.Body.DragId,
		wsData.Body.DropId,
	)
	if err != nil {
		_data = s2cReorderSub{
			Head: struct {
				Cmd s2cCmds "json:\"cmd\""
			}{
				Cmd: s2cCmdReorderSubBack,
			},
			Body: struct {
				OperationUser string "json:\"operation_user\""
				Status        bool   "json:\"status\""
				DragId        uint   "json:\"drag_id\""
				DropId        uint   "json:\"drop_id\""
			}{
				OperationUser: wsData.Body.OperationUser,
				Status:        false,
				DragId:        wsData.Body.DragId,
				DropId:        wsData.Body.DropId,
			},
		}
	} else {
		_data = s2cReorderSub{
			Head: struct {
				Cmd s2cCmds "json:\"cmd\""
			}{
				Cmd: s2cCmdReorderSubBack,
			},
			Body: struct {
				OperationUser string "json:\"operation_user\""
				Status        bool   "json:\"status\""
				DragId        uint   "json:\"drag_id\""
				DropId        uint   "json:\"drop_id\""
			}{
				OperationUser: wsData.Body.OperationUser,
				Status:        true,
				DragId:        wsData.Body.DragId,
				DropId:        wsData.Body.DropId,
			},
		}
	}

	data, marshalErr := json.Marshal(&_data)
	if marshalErr != nil {
		return marshalErr
	}
	m.data = data

	WsHub.broadcast <- *m

	return nil
}

func (m *message) handleSendSubtitle() error {
	var wsData c2sSendSubtitle
	unmarshalErr := json.Unmarshal(m.data, &wsData)
	if unmarshalErr != nil {
		return unmarshalErr
	}
	var _data s2cSendSubtitle
	err := db.SendSubtitle(wsData.Body.Subtitle)
	if err != nil {
		_data = s2cSendSubtitle{
			Head: struct {
				Cmd s2cCmds "json:\"cmd\""
			}{
				Cmd: s2cCmdSendSubtitle,
			},
			Body: struct {
				Status   bool           "json:\"status\""
				Subtitle model.Subtitle "json:\"subtitle\""
			}{
				Status:   false,
				Subtitle: wsData.Body.Subtitle,
			},
		}
	} else {
		_data = s2cSendSubtitle{
			Head: struct {
				Cmd s2cCmds "json:\"cmd\""
			}{
				Cmd: s2cCmdSendSubtitle,
			},
			Body: struct {
				Status   bool           "json:\"status\""
				Subtitle model.Subtitle "json:\"subtitle\""
			}{
				Status:   true,
				Subtitle: wsData.Body.Subtitle,
			},
		}
	}

	data, marshalErr := json.Marshal(&_data)
	if marshalErr != nil {
		return marshalErr
	}
	m.data = data

	WsHub.broadcast <- *m

	ctx := context.Background()
	setErr := setNowSubtitleData(ctx, m.room, wsData.Body.Subtitle)
	if setErr != nil {
		logger.Warn("ws handler", fmt.Sprintf("set subtitle to redis: %v", setErr))
	}
	return nil
}

func (m *message) handleSendSubtitleDirect() error {
	var wsData c2sSendSubtitleDirect
	unmarshalErr := json.Unmarshal(m.data, &wsData)
	if unmarshalErr != nil {
		return unmarshalErr
	}
	var _data s2cSendSubtitle
	sub, err := db.DirectSendSubtitle(wsData.Body.Subtitle)
	if err != nil {
		_data = s2cSendSubtitle{
			Head: struct {
				Cmd s2cCmds "json:\"cmd\""
			}{
				Cmd: s2cCmdSendSubtitleDirect,
			},
			Body: struct {
				Status   bool           "json:\"status\""
				Subtitle model.Subtitle "json:\"subtitle\""
			}{
				Status:   false,
				Subtitle: sub,
			},
		}
	} else {
		_data = s2cSendSubtitle{
			Head: struct {
				Cmd s2cCmds "json:\"cmd\""
			}{
				Cmd: s2cCmdSendSubtitleDirect,
			},
			Body: struct {
				Status   bool           "json:\"status\""
				Subtitle model.Subtitle "json:\"subtitle\""
			}{
				Status:   true,
				Subtitle: sub,
			},
		}
	}

	data, marshalErr := json.Marshal(&_data)
	if marshalErr != nil {
		return marshalErr
	}
	m.data = data

	WsHub.broadcast <- *m

	ctx := context.Background()
	setErr := setNowSubtitleData(ctx, m.room, wsData.Body.Subtitle)
	if setErr != nil {
		logger.Warn("ws handler", fmt.Sprintf("set subtitle to redis: %v", setErr))
	}
	return nil
}

func (m *message) handleChangeStyle() error {
	var wsData c2sChangeStyle
	unmarshalErr := json.Unmarshal(m.data, &wsData)
	if unmarshalErr != nil {
		return unmarshalErr
	}
	_data := s2cChangeStyle{
		Head: struct {
			Cmd s2cCmds "json:\"cmd\""
		}{
			Cmd: s2cCmdChangeStyle,
		},
		Body: wsData.Body,
	}

	data, marshalErr := json.Marshal(&_data)
	if marshalErr != nil {
		return marshalErr
	}
	m.data = data

	WsHub.broadcast <- *m

	ctx := context.Background()
	setErr := setNowStyleData(ctx, m.room, wsData.Body)
	if setErr != nil {
		logger.Warn("ws handler", fmt.Sprintf("set style to redis: %v", setErr))
	}
	return nil
}

// set now style data to redis by wsroom
func setNowStyleData(ctx context.Context, wsroom string, style ChangeStyleBody) error {
	rdbKey := stat.MakeStyleRdbKey(wsroom)
	rdbValueStr, marshalErr := json.MarshalToString(style)
	if marshalErr != nil {
		return marshalErr
	}
	rdbErr := db.Rdb.Set(ctx, rdbKey, rdbValueStr, 24*time.Hour).Err()
	if rdbErr != nil {
		return rdbErr
	}
	logger.Info("wsHandler", "set now style data successful")
	return nil
}

// set now subtitle data to redis by wsroom
func setNowSubtitleData(ctx context.Context, wsroom string, s model.Subtitle) error {
	rdbKey := stat.MakeSubtitleRdbKey(wsroom)
	rdbValueStr, marshalErr := json.MarshalToString(s)
	if marshalErr != nil {
		return marshalErr
	}
	rdbErr := db.Rdb.Set(ctx, rdbKey, rdbValueStr, 24*time.Hour).Err()
	if rdbErr != nil {
		return rdbErr
	}
	logger.Info("wsHandler", "set now subtitle data successful")
	return nil
}

// get now style 利用 changeStyle 的接口发送style给client
func (m *message) handleGetNowRoomStyle() error {
	var style ChangeStyleBody
	ctx := context.Background()

	val, rdbErr := db.Rdb.Get(ctx, stat.MakeStyleRdbKey(m.room)).Result()
	if rdbErr == redis.Nil {
		logger.Warn("wsHandler", "redis get now style, the key is expire or undefined")
		// 不存在key可能是正常行为, 比如第一次发送
		// 但是为了不覆盖client的默认style设置所以这里返回一个nil以停止函数
		return nil
	} else if rdbErr != nil {
		return rdbErr
	} else if val == "" {
		logger.Warn("wsHandler", "redis get now style, the value is empty")
	} else {
		unmarshalErr := json.UnmarshalFromString(val, &style)
		if unmarshalErr != nil {
			return unmarshalErr
		}
	}

	_data := s2cChangeStyle{
		Head: struct {
			Cmd s2cCmds "json:\"cmd\""
		}{
			Cmd: s2cCmdChangeStyle,
		},
		Body: style,
	}
	data, marshalErr := json.Marshal(&_data)
	if marshalErr != nil {
		return marshalErr
	}
	m.data = data

	WsHub.castself <- *m

	return nil
}

// get now subtitle 利用 sendSubtitle 的接口发送当前sub给client
func (m *message) handleGetNowRoomSub() error {
	var sub model.Subtitle
	ctx := context.Background()

	val, rdbErr := db.Rdb.Get(ctx, stat.MakeSubtitleRdbKey(m.room)).Result()
	if rdbErr == redis.Nil {
		logger.Warn("wsHandler", "redis get now room sub, the key is expire or undefined")
		// subtitle和style不同, 如果不存在则发送一个空白给client以清空当前
	} else if rdbErr != nil {
		return rdbErr
	} else if val == "" {
		logger.Warn("wsHandler", "redis get now room sub, the value is empty")
	} else {
		unmarshalErr := json.UnmarshalFromString(val, &sub)
		if unmarshalErr != nil {
			return unmarshalErr
		}
	}

	_data := s2cSendSubtitle{
		Head: struct {
			Cmd s2cCmds "json:\"cmd\""
		}{
			Cmd: s2cCmdSendSubtitle,
		},
		Body: struct {
			Status   bool           "json:\"status\""
			Subtitle model.Subtitle "json:\"subtitle\""
		}{
			Status:   true,
			Subtitle: sub,
		},
	}
	data, marshalErr := json.Marshal(&_data)
	if marshalErr != nil {
		return marshalErr
	}
	m.data = data

	WsHub.castself <- *m

	return nil
}

func (m *message) handleBatchAddSubs() error {
	var wsData c2sBatchAddSubs
	unmarshalErr := json.Unmarshal(m.data, &wsData)
	if unmarshalErr != nil {
		return unmarshalErr
	}

	var _data s2cBatchAddSubs

	dbErr := db.BatchAddSubs(wsData.Body.Subtitles)
	if dbErr != nil {
		_data = s2cBatchAddSubs{
			Head: struct {
				Cmd s2cCmds "json:\"cmd\""
			}{
				Cmd: s2cCmdBatchAddSubs,
			},
			Body: struct {
				Status bool "json:\"status\""
			}{
				Status: false,
			},
		}
	} else {
		_data = s2cBatchAddSubs{
			Head: struct {
				Cmd s2cCmds "json:\"cmd\""
			}{
				Cmd: s2cCmdBatchAddSubs,
			},
			Body: struct {
				Status bool "json:\"status\""
			}{
				Status: true,
			},
		}
	}

	data, marshalErr := json.Marshal(&_data)
	if marshalErr != nil {
		return marshalErr
	}
	m.data = data

	WsHub.castself <- *m

	return nil
}

// 以后可能会利用心跳做点什么,但目前用不到这个代码
func (m *message) handleHeartBeat() error {
	var wsData c2sHeartBeat
	unmarshalErr := json.Unmarshal(m.data, &wsData)
	if unmarshalErr != nil {
		return unmarshalErr
	}
	var dbRoomType uint
	if wsData.Body.RoomType == auto {
		dbRoomType = 2
	} else if wsData.Body.RoomType == nomal {
		dbRoomType = 1
	} else {
		return errors.New("room type must be auto or nomal")
	}

	err := db.CheckWsroomType(dbRoomType, wsData.Body.RoomId)

	if err != nil {
		return err
	}

	_data := s2cHeartBeat{
		Head: struct {
			Cmd s2cCmds "json:\"cmd\""
		}{
			Cmd: s2cCmdHeartBeat,
		},
		Body: struct {
			Data interface{} "json:\"data\""
		}{
			Data: "heartBeat",
		},
	}
	data, marshalErr := json.Marshal(&_data)
	if marshalErr != nil {
		return marshalErr
	}
	m.data = data
	return nil
}
