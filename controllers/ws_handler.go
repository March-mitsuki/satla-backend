package controllers

import (
	"errors"

	"github.com/March-mitsuki/satla-backend/controllers/db"
	"github.com/March-mitsuki/satla-backend/model"
)

func (rUser *roomUsers) addUser(wsroom, uname string) {
	// 如果存在该房间则直接追加new user, 若不存在则创建一个新房间并追加
	_, ok := (*rUser)[wsroom]
	if !ok {
		(*rUser)[wsroom] = []string{uname}
		return
	}
	(*rUser)[wsroom] = append((*rUser)[wsroom], uname)
	return
}

func (rUser *roomUsers) delUser(wsroom, uname string) error {
	// 删除指定room内的指定user, 若删除后房间内不存在user, 则会连房间一起删除
	// 若不存在该房间或传入用户名为空值则返回一个error
	if uname == "" {
		return errors.New("user name is empty")
	}
	_, ok := (*rUser)[wsroom]
	if !ok {
		return errors.New("no such room")
	} else {
		for idx, v := range (*rUser)[wsroom] {
			if v == uname {
				// 如果房间内不存在该用户则不会触发删除逻辑, 函数遍历后返回nil
				(*rUser)[wsroom] = append((*rUser)[wsroom][:idx], (*rUser)[wsroom][idx+1:]...)
				break
			}
		}
		if length := len((*rUser)[wsroom]); length == 0 {
			delete(*rUser, wsroom)
		}
		return nil
	}
}

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
	return nil
}

func (m *message) handleAddTranslatedSub() error {
	var wsData c2sAddTranslatedSub
	unmarshalErr := json.Unmarshal(m.data, &wsData)
	if unmarshalErr != nil {
		return unmarshalErr
	}
	// 这里收到的subtitle的id和project_id为0, 需要额外操作
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
		Body: struct {
			Reversed bool   "json:\"reverse\""
			Subtitle string "json:\"subtitle\""
			Origin   string "json:\"origin\""
		}{
			Subtitle: wsData.Body.Subtitle,
			Origin:   wsData.Body.Origin,
		},
	}

	data, marshalErr := json.Marshal(&_data)
	if marshalErr != nil {
		return marshalErr
	}
	m.data = data
	return nil
}

func (m *message) handleChangeBilingual() error {
	var wsData c2sChangeBilingual
	unmarshalErr := json.Unmarshal(m.data, &wsData)
	if unmarshalErr != nil {
		return unmarshalErr
	}
	_data := s2cChangeBilingual{
		Head: struct {
			Cmd s2cCmds "json:\"cmd\""
		}{
			Cmd: s2cCmdChangeBilingual,
		},
		Body: struct {
			Bilingual bool "json:\"bilingual\""
		}{
			Bilingual: wsData.Body.Bilingual,
		},
	}

	data, marshalErr := json.Marshal(&_data)
	if marshalErr != nil {
		return marshalErr
	}
	m.data = data
	return nil
}

func (m *message) handleChangeReversed() error {
	var wsData c2sChangeReversed
	unmarshalErr := json.Unmarshal(m.data, &wsData)
	if unmarshalErr != nil {
		return unmarshalErr
	}
	_data := s2cChangeReversed{
		Head: struct {
			Cmd s2cCmds "json:\"cmd\""
		}{
			Cmd: s2cCmdChangeReversed,
		},
		Body: struct {
			Reversed bool "json:\"reversed\""
		}{
			Reversed: wsData.Body.Reversed,
		},
	}

	data, marshalErr := json.Marshal(&_data)
	if marshalErr != nil {
		return marshalErr
	}
	m.data = data
	return nil
}

// 以后可能会利用心跳做点什么,但目前用不到这个代码
func (m *message) handleHeartBeat() error {
	var wsData c2sHeartBeat
	unmarshalErr := json.Unmarshal(m.data, &wsData)
	if unmarshalErr != nil {
		return unmarshalErr
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
