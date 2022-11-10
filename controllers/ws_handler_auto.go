package controllers

import (
	"context"
	"fmt"
	"sync"
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

func (m *message) handlePlayStart(ctx context.Context) error {
	var wsData c2sPlayStart
	unmarshalErr := json.Unmarshal(m.data, &wsData)
	if unmarshalErr != nil {
		logger.Err(fmt.Sprintf("unmarshal wsData === \n %v \n ===", unmarshalErr))
		return unmarshalErr
	}
	var autoSubs []model.AutoSub
	result := db.Mdb.Where("list_id = ?", wsData.Body.ListId).Find(&autoSubs)
	if result.Error != nil {
		logger.Err(fmt.Sprintf("[auto] === \n %v \n ===", result.Error))
		broadcastAutoPlayErr(m, "start err")
		return result.Error
	}
	go autoPlayStart(m, ctx, autoSubs)
	return nil
}

func (m *message) handlePlayEnd(endPlay context.CancelFunc) error {
	endPlay()
	return nil
}

// return a s2cAutoChangeSub struct
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

// main logic of auto play
func autoPlayStart(m *message, ctx context.Context, autoSubs []model.AutoSub) {
	defer func() {
		logger.Info("Auto Play Finish")
	}()
	logger.Info("auto play start")
	for _, v := range autoSubs {
		_data := makeAutoChangeSub(v)
		data, marshalErr := json.Marshal(&_data)
		if marshalErr != nil {
			logger.Err(fmt.Sprintf("[auto] === \n %v \n ===", marshalErr))
			broadcastAutoPlayErr(m, "loop marshal error")
			return
		}
		m.data = data
		d, dErr := time.ParseDuration(fmt.Sprintf("%vs", v.Duration))
		if dErr != nil {
			logger.Err(fmt.Sprintf("[auto] === \n %v \n ===", dErr))
			broadcastAutoPlayErr(m, "loop parse duration error")
			return
		}
		WsHub.broadcast <- *m
		var wg sync.WaitGroup
		timer := time.NewTimer(d)
		wg.Add(1)
		go cancelableSleep(ctx, timer, &wg)
		wg.Wait()
	}
	return
}

func cancelableSleep(ctx context.Context, t *time.Timer, wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
		logger.Info("Sleep done")
	}()
LOOP:
	for {
		select {
		case <-ctx.Done():
			break LOOP
		case <-t.C:
			break LOOP
		}
	}
	return
}

func broadcastAutoPlayErr(m *message, reason string) *message {
	_data := s2cAutoPlayErr{
		Head: struct {
			Cmd s2cCmds "json:\"cmd\""
		}{
			Cmd: s2cCmdAutoPlayErr,
		},
		Body: struct {
			Msg string "json:\"msg\""
		}{
			Msg: reason,
		},
	}
	data, marshalErr := json.Marshal(&_data)
	if marshalErr != nil {
		logger.Err(fmt.Sprintf("make auto play err: %v \n", marshalErr))
		return nil
	}
	m.data = data
	WsHub.broadcast <- *m
	return m
}
