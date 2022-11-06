package controllers

import (
	"github.com/March-mitsuki/satla-backend/controllers/db"
	"github.com/March-mitsuki/satla-backend/model"
)

func (m *message) handleGetRoomAutoLists() error {
	var wsData c2sGetAutoLists
	unmarshalErr := json.Unmarshal(m.data, &wsData)
	if unmarshalErr != nil {
		return unmarshalErr
	}
	autoLists, err := db.GetRoomAutoLists(wsData.Body.RoomId)
	var _data s2cGetAutoLists
	if err != nil {
		_data = s2cGetAutoLists{
			Head: struct {
				Cmd s2cCmds "json:\"cmd\""
			}{
				Cmd: s2cCmdGetAutoLists,
			},
			Body: struct {
				Status    bool             "json:\"status\""
				AutoLists []model.AutoList "json:\"auto_lists\""
			}{
				Status:    false,
				AutoLists: autoLists,
			},
		}
	} else {
		_data = s2cGetAutoLists{
			Head: struct {
				Cmd s2cCmds "json:\"cmd\""
			}{
				Cmd: s2cCmdGetAutoLists,
			},
			Body: struct {
				Status    bool             "json:\"status\""
				AutoLists []model.AutoList "json:\"auto_lists\""
			}{
				Status:    true,
				AutoLists: autoLists,
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

func (m *message) handleAddAutoSub() error {
	var wsData c2sAddAutoSub
	unmarshalErr := json.Unmarshal(m.data, &wsData)
	if unmarshalErr != nil {
		return unmarshalErr
	}
	arg := db.ArgAddAutoSub{
		AutoSubs: wsData.Body.AutoSubs,
		Memo:     wsData.Body.Memo,
	}
	newList, err := db.AddAutoSub(arg)

	var _data s2cAddAutoSub
	if err != nil {
		_data = s2cAddAutoSub{
			Head: struct {
				Cmd s2cCmds "json:\"cmd\""
			}{
				Cmd: s2cCmdAddAutoSub,
			},
			Body: struct {
				Status  bool           "json:\"status\""
				NewList model.AutoList "json:\"new_list\""
			}{
				Status:  false,
				NewList: newList,
			},
		}
	} else {
		_data = s2cAddAutoSub{
			Head: struct {
				Cmd s2cCmds "json:\"cmd\""
			}{
				Cmd: s2cCmdAddAutoSub,
			},
			Body: struct {
				Status  bool           "json:\"status\""
				NewList model.AutoList "json:\"new_list\""
			}{
				Status:  true,
				NewList: newList,
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
