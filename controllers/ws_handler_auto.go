package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/March-mitsuki/satla-backend/utils/logger"

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

func makeAutoChangeSub(s model.AutoSub) s2cAutoChangeSub {
	data := s2cAutoChangeSub{
		Head: struct {
			Cmd s2cCmds "json:\"cmd\""
		}{
			Cmd: s2cCmdAutoChangeSub,
		},
		Body: struct {
			AutoSub model.AutoSub "json:\"auto_sub\""
		}{
			AutoSub: s,
		},
	}
	return data
}

func AutoPlayStart(m message) {
	logger.Info("auto play start")
	var wsData c2sPlayStart
	unmarshalErr := json.Unmarshal(m.data, &wsData)
	if unmarshalErr != nil {
		logger.Err(fmt.Sprintf("unmarshal wsData ERROR === \n %v \n ===", unmarshalErr))
		return
	}
	var autoSubs []model.AutoSub
	result := db.Mdb.Where("list_id = ?", wsData.Body.ListId).Find(&autoSubs)
	if result.Error != nil {
		logger.Err(fmt.Sprintf("[auto] ERROR === \n %v \n ===", result.Error))
		return
	}
	for _, v := range autoSubs {
		_data := makeAutoChangeSub(v)
		data, marshalErr := json.Marshal(&_data)
		if marshalErr != nil {
			logger.Err(fmt.Sprintf("[auto] ERROR === \n %v \n ===", result.Error))
			return
		}
		(&m).data = data
		d, dErr := time.ParseDuration(fmt.Sprintf("%vs", v.Duration))
		if dErr != nil {
			logger.Err(fmt.Sprintf("[auto] ERROR === \n %v \n ===", result.Error))
			return
		}
		WsHub.broadcast <- m
		time.Sleep(d)
	}
	return
}

func cancelableSleep() {
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		fmt.Println("sleep done")
		cancel()
		return
	}()
	go func() {
		t := time.Now()
		select {
		case <-ctx.Done():
		case <-time.After(2 * time.Second):
		}
		fmt.Printf("here after: %v\n", time.Since(t))
	}()
	cancel()
	time.Sleep(3 * time.Second)
	fmt.Println("done")
}
