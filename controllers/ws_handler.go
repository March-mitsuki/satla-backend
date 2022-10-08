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
	var wsData c2sAddUser
	unmarshalErr := json.Unmarshal(m.data, &wsData)
	if unmarshalErr != nil {
		return "", unmarshalErr
	}
	// fmt.Printf("\n --parse add user-- \n %+v \n", wsData)
	allRoomUsers.addUser(m.room, wsData.Body.Uname)
	_data := s2cAddUser{
		Head: struct {
			Cmd string "json:\"cmd\""
		}{
			Cmd: s2cCmdAddUser,
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
	subtitles, dbGetErr := db.GetRoomSubtitles(wsData.Body.Roomid)
	if dbGetErr != nil {
		return dbGetErr
	}
	_data := s2cGetRoomSubtitles{
		Head: struct {
			Cmd string "json:\"cmd\""
		}{
			Cmd: s2cCmdGetRoomSubtitles,
		},
		Body: struct {
			Subtitles []model.Subtitle "json:\"subtitles\""
		}{
			Subtitles: subtitles,
		},
	}
	data, marshalErr := json.Marshal(&_data)
	if marshalErr != nil {
		return marshalErr
	}
	m.data = data
	return nil
}
