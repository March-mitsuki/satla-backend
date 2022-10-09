package controllers

import (
	"errors"
	"vvvorld/controllers/db"
	"vvvorld/model"
)

func (rUser *roomUsers) addUser(roomid, uname string) {
	// 如果存在该房间则直接追加new user, 若不存在则创建一个新房间并追加
	_, ok := (*rUser)[roomid]
	if !ok {
		(*rUser)[roomid] = []string{uname}
		return
	}
	(*rUser)[roomid] = append((*rUser)[roomid], uname)
	return
}

func (rUser *roomUsers) delUser(roomid, uname string) error {
	// 删除指定room内的指定user, 若删除后房间内不存在user, 则会连房间一起删除
	// 若不存在该房间则返回一个error
	_, ok := (*rUser)[roomid]
	if !ok {
		return errors.New("no such room")
	} else {
		for idx, v := range (*rUser)[roomid] {
			if v == uname {
				(*rUser)[roomid] = append((*rUser)[roomid][:idx], (*rUser)[roomid][idx+1:]...)
				break
			}
		}
		if length := len((*rUser)[roomid]); length == 0 {
			delete(*rUser, roomid)
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
	subtitles, order, dbGetErr := db.GetRoomSubtitles(wsData.Body.Roomid)
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
		ProjectId:     wsData.Body.ProjectId,
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
			ProjectId      uint   "json:\"project_id\""
			NewSubtitleId  uint   "json:\"new_subtitle_id\""
			PreSubtitleIdx uint   "json:\"pre_subtitle_idx\""
			CheckedBy      string "json:\"checked_by\""
		}{
			ProjectId:      wsData.Body.PreSubtitleId,
			PreSubtitleIdx: wsData.Body.PreSubtitleIdx,
			NewSubtitleId:  newSubId,
			CheckedBy:      wsData.Body.CheckedBy,
		},
	}
	data, marshalErr := json.Marshal(&_data)
	if marshalErr != nil {
		return nil
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
		ProjectId:     wsData.Body.ProjectId,
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
			ProjectId      uint   "json:\"project_id\""
			NewSubtitleId  uint   "json:\"new_subtitle_id\""
			PreSubtitleIdx uint   "json:\"pre_subtitle_idx\""
			CheckedBy      string "json:\"checked_by\""
		}{
			ProjectId:      wsData.Body.PreSubtitleId,
			PreSubtitleIdx: wsData.Body.PreSubtitleIdx,
			NewSubtitleId:  newSubId,
			CheckedBy:      wsData.Body.CheckedBy,
		},
	}
	data, marshalErr := json.Marshal(&_data)
	if marshalErr != nil {
		return nil
	}
	m.data = data
	return nil
}
