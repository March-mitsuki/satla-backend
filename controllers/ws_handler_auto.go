package controllers

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/March-mitsuki/satla-backend/utils/logger"
	"github.com/March-mitsuki/satla-backend/utils/stat"
	"github.com/go-redis/redis/v9"

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

	WsHub.castself <- *m

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

	WsHub.broadcast <- *m

	return nil
}

func (m *message) handlePlayStart(ctx context.Context, ope chan autoOpeData) error {
	var wsData c2sPlayStart
	unmarshalErr := json.Unmarshal(m.data, &wsData)
	if unmarshalErr != nil {
		logger.Err("wsHandler", fmt.Sprintf("unmarshal wsData === \n %v \n ===", unmarshalErr))
		return unmarshalErr
	}
	_data := s2cAutoPlayStart{
		Head: struct {
			Cmd s2cCmds "json:\"cmd\""
		}{
			Cmd: s2cCmdAutoPlayStart,
		},
		Body: struct {
			ListId uint "json:\"list_id\""
		}{
			ListId: wsData.Body.ListId,
		},
	}
	data, marshalErr := json.Marshal(&_data)
	if marshalErr != nil {
		return marshalErr
	}
	m.data = data
	WsHub.broadcast <- *m

	autoSubs, dbErr := db.HandleAutoPlayStart(wsData.Body.ListId)
	if dbErr != nil {
		logger.Err("wsHandler", fmt.Sprintf("[auto] === \n %v \n ===", dbErr))
		broadcastAutoPlayErr(m, "start err")
		return dbErr
	}
	go autoPlayStart(ctx, m, &autoSubs, ope)
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

func calc() calcData {
	num := 0
	sum := func(i int) int {
		num += i
		return num
	}
	sub := func(i int) int {
		num -= i
		return num
	}
	return calcData{
		adder: sum,
		suber: sub,
	}
}

type calcData struct {
	adder func(int) int
	suber func(int) int
}

func autoPlayStart(
	ctx context.Context,
	m *message,
	autoSubs *[]model.AutoSub,
	ope chan autoOpeData,
) {
	logger.Info("autoPlay", "auto play whiout loop start")
	ca := calc()
	subtitle := (*autoSubs)[0]
	d, dErr := time.ParseDuration(fmt.Sprintf("%vs", subtitle.Duration))
	if dErr != nil {
		logger.Err("autoPlay", fmt.Sprintf("[auto] === \n %v \n ===", dErr))
		broadcastAutoPlayErr(m, "loop start parse duration error")
		return
	}
	broadcastAutoChangeSub(m, &subtitle)
	broadcastAutoPreviewChange(m, autoSubs, 0)

	// set redis for room state checker
	rdbKey := stat.MakeAutoRdbKey(m.room)
	rdbValue := autoPlayState{
		Wsroom:  m.room,
		State:   playing,
		ListId:  subtitle.ListId,
		NowSub:  subtitle,
		Preview: makeAutoPreview(autoSubs, 0),
	}
	rdbValueStr, marshalErr := json.MarshalToString(rdbValue)
	if marshalErr != nil {
		logger.Err("json", fmt.Sprintf("auto play start json marshal: %v", marshalErr))
		return
	}
	rdbErr := db.Rdb.Set(ctx, rdbKey, rdbValueStr, 5*time.Minute).Err()
	if rdbErr != nil {
		logger.Err("rdb", fmt.Sprintf("auto play start set: %v", rdbErr))
		broadcastAutoPlayErr(m, "rdb err")
		return
	}

	timer := time.NewTimer(d)
	var wg sync.WaitGroup
	wg.Add(1)
	cancelableSleep(ctx, timer, &wg, ca, m, autoSubs, ope)
	wg.Wait()
}

func autoPlayLoop(
	ctx context.Context,
	ca calcData,
	m *message,
	autoSubs *[]model.AutoSub,
	wg *sync.WaitGroup,
	ope chan autoOpeData,
) {
	logger.Info("autoPlay", "auto play loop called")
	if ca.adder(0) >= len(*autoSubs)-1 {
		// 结束有两个逻辑, 一个是这里的自动结束, 还有一个在cancelableSleep的ctx.Done()里
		go broadcastSendBlank(m)
		go broadcastPreviewEnd(m)
		go broadcastAutoPlayEnd(m)
		allAutoCtxs.delCtx(m.room, (*autoSubs)[0].ListId)

		// set redis for room state checker
		rdbKey := stat.MakeAutoRdbKey(m.room)
		rdbValue := autoPlayState{
			Wsroom: m.room,
			State:  stopped,
		}
		rdbValueStr, marshalErr := json.MarshalToString(rdbValue)
		if marshalErr != nil {
			logger.Err("json", fmt.Sprintf("auto play stop json marshal: %v", marshalErr))
			return
		}
		rdbErr := db.Rdb.Set(ctx, rdbKey, rdbValueStr, 5*time.Minute).Err()
		if rdbErr != nil {
			logger.Err("rdb", fmt.Sprintf("auto play stop set: %v", rdbErr))
			return
		}
		return
	}
	subtitle := (*autoSubs)[ca.adder(1)]
	d, dErr := time.ParseDuration(fmt.Sprintf("%vs", subtitle.Duration))
	if dErr != nil {
		logger.Err("autoPlay", fmt.Sprintf("[auto] === \n %v \n ===", dErr))
		broadcastAutoPlayErr(m, "loop parse duration error")
		return
	}
	broadcastAutoChangeSub(m, &subtitle)
	broadcastAutoPreviewChange(m, autoSubs, ca.adder(0))

	// set redis for room state checker
	rdbKey := stat.MakeAutoRdbKey(m.room)
	rdbValue := autoPlayState{
		Wsroom:  m.room,
		State:   playing,
		ListId:  subtitle.ListId,
		NowSub:  subtitle,
		Preview: makeAutoPreview(autoSubs, ca.adder(0)),
	}
	rdbValueStr, marshalErr := json.MarshalToString(rdbValue)
	if marshalErr != nil {
		logger.Err("json", fmt.Sprintf("auto play loop json marshal: %v", marshalErr))
		return
	}
	rdbErr := db.Rdb.Set(ctx, rdbKey, rdbValueStr, 5*time.Minute).Err()
	if rdbErr != nil {
		logger.Err("rdb", fmt.Sprintf("auto play loop set: %v", rdbErr))
		return
	}

	timer := time.NewTimer(d)
	wg.Add(1)
	cancelableSleep(ctx, timer, wg, ca, m, autoSubs, ope)
	wg.Wait()
}

func broadcastAutoChangeSub(m *message, sub *model.AutoSub) {
	_data := makeAutoChangeSub(*sub)
	data, marshalErr := json.Marshal(&_data)
	if marshalErr != nil {
		logger.Err("autoPlay", fmt.Sprintf("[auto] === \n %v \n ===", marshalErr))
		broadcastAutoPlayErr(m, "loop marshal error")
		return
	}
	m.data = data
	WsHub.broadcast <- *m
}

func broadcastAutoPreviewChange(m *message, autoSubs *[]model.AutoSub, nowNum int) {
	_data := s2cAutoPreviewChange{
		Head: struct {
			Cmd s2cCmds "json:\"cmd\""
		}{
			Cmd: s2cCmdAutoPreviewChange,
		},
		Body: makeAutoPreview(autoSubs, nowNum),
	}

	data, marshalErr := json.Marshal(&_data)
	if marshalErr != nil {
		logger.Err("autoPlay", fmt.Sprintf("[auto] === \n %v \n ===", marshalErr))
		broadcastAutoPlayErr(m, "loop marshal error")
		return
	}
	m.data = data
	WsHub.broadcast <- *m
}

func makeAutoPreview(autoSubs *[]model.AutoSub, nowNum int) autoPreview {
	var result autoPreview
	if nowNum-1 < 0 {
		result = autoPreview{
			BehindTwo: model.AutoSub{},
			Behind:    model.AutoSub{},
			Main:      (*autoSubs)[nowNum],
			Next:      (*autoSubs)[nowNum+1],
			NextTwo:   (*autoSubs)[nowNum+2],
		}
	} else if nowNum-2 < 0 {
		result = autoPreview{
			BehindTwo: model.AutoSub{},
			Behind:    (*autoSubs)[nowNum-1],
			Main:      (*autoSubs)[nowNum],
			Next:      (*autoSubs)[nowNum+1],
			NextTwo:   (*autoSubs)[nowNum+2],
		}
	} else if nowNum+1 > len(*autoSubs)-1 {
		result = autoPreview{
			BehindTwo: (*autoSubs)[nowNum-2],
			Behind:    (*autoSubs)[nowNum-1],
			Main:      (*autoSubs)[nowNum],
			Next:      model.AutoSub{},
			NextTwo:   model.AutoSub{},
		}
	} else if nowNum+2 > len(*autoSubs)-1 {
		result = autoPreview{
			BehindTwo: (*autoSubs)[nowNum-2],
			Behind:    (*autoSubs)[nowNum-1],
			Main:      (*autoSubs)[nowNum],
			Next:      (*autoSubs)[nowNum+1],
			NextTwo:   model.AutoSub{},
		}
	} else {
		result = autoPreview{
			BehindTwo: (*autoSubs)[nowNum-2],
			Behind:    (*autoSubs)[nowNum-1],
			Main:      (*autoSubs)[nowNum],
			Next:      (*autoSubs)[nowNum+1],
			NextTwo:   (*autoSubs)[nowNum+2],
		}
	}
	return result
}

func cancelableSleep(
	ctx context.Context,
	t *time.Timer,
	wg *sync.WaitGroup,
	ca calcData,
	m *message,
	autoSubs *[]model.AutoSub,
	ope chan autoOpeData,
) {
	defer func() {
		wg.Done()
		logger.Info("autoPlay", "loop sleep done")
	}()
LOOP:
	for {
		select {
		case <-ctx.Done():
			t.Stop()
			go broadcastSendBlank(m)
			go broadcastPreviewEnd(m)
			go broadcastAutoPlayEnd(m)
			allAutoCtxs.delCtx(m.room, (*autoSubs)[0].ListId)
			// set redis for room state checker
			rdbKey := stat.MakeAutoRdbKey(m.room)
			rdbValue := autoPlayState{
				Wsroom: m.room,
				State:  stopped,
			}
			rdbValueStr, marshalErr := json.MarshalToString(rdbValue)
			if marshalErr != nil {
				logger.Err("json", fmt.Sprintf("auto play stop manually json marshal: %v", marshalErr))
				return
			}
			rdbCtx := context.Background()
			rdbErr := db.Rdb.Set(rdbCtx, rdbKey, rdbValueStr, 5*time.Minute).Err()
			if rdbErr != nil {
				logger.Err("rdb", fmt.Sprintf("auto play stop manually set: %v", rdbErr))
				return
			}
			break LOOP
		case o := <-ope:
			switch o.opeType {
			case foward:
				t.Stop()
				go autoPlayLoop(ctx, ca, m, autoSubs, wg, ope)

			case fowardTwice:
				t.Stop()
				ca.adder(1)
				go autoPlayLoop(ctx, ca, m, autoSubs, wg, ope)

			case rewind:
				t.Stop()
				ca.suber(2)
				go autoPlayLoop(ctx, ca, m, autoSubs, wg, ope)

			case rewindTwice:
				t.Stop()
				ca.suber(3)
				go autoPlayLoop(ctx, ca, m, autoSubs, wg, ope)

			case pause:
				logger.Info("sleep", "pause now")
				t.Stop()
				// set redis for room state checker
				rdbKey := stat.MakeAutoRdbKey(m.room)
				rdbValue := autoPlayState{
					Wsroom:  m.room,
					State:   playing,
					ListId:  (*autoSubs)[0].ListId,
					NowSub:  (*autoSubs)[ca.adder(0)],
					Preview: makeAutoPreview(autoSubs, ca.adder(0)),
				}
				rdbValueStr, marshalErr := json.MarshalToString(rdbValue)
				if marshalErr != nil {
					logger.Err("json", fmt.Sprintf("auto play stop manually json marshal: %v", marshalErr))
					return
				}
				rdbErr := db.Rdb.Set(ctx, rdbKey, rdbValueStr, 5*time.Minute).Err()
				if rdbErr != nil {
					logger.Err("rdb", fmt.Sprintf("auto play pause set: %v", rdbErr))
					return
				}
				for {
					o := <-ope
					if o.opeType == restart {
						go autoPlayLoop(ctx, ca, m, autoSubs, wg, ope)
						break
					}
				}
			}
			logger.Info("sleep", "case done")
			break LOOP
		case <-t.C:
			go autoPlayLoop(ctx, ca, m, autoSubs, wg, ope)
			break LOOP
		}
	}
	return
}

func broadcastSendBlank(m *message) {
	_lastSend := makeAutoChangeSub(model.AutoSub{})
	data, marshalErr := json.Marshal(&_lastSend)
	if marshalErr != nil {
		logger.Err("autoPlay", fmt.Sprintf("[auto] === \n %v \n ===", marshalErr))
		broadcastAutoPlayErr(m, "loop marshal error")
		return
	}
	m.data = data
	WsHub.broadcast <- *m
	logger.Info("autoPlay", "last send")
	return
}

func broadcastPreviewEnd(m *message) {
	_data := s2cAutoPreviewChange{
		Head: struct {
			Cmd s2cCmds "json:\"cmd\""
		}{
			Cmd: s2cCmdAutoPreviewChange,
		},
		Body: struct {
			BehindTwo model.AutoSub "json:\"behind_two\""
			Behind    model.AutoSub "json:\"behind\""
			Main      model.AutoSub "json:\"main\""
			Next      model.AutoSub "json:\"next\""
			NextTwo   model.AutoSub "json:\"next_two\""
		}{
			BehindTwo: model.AutoSub{},
			Behind:    model.AutoSub{},
			Main:      model.AutoSub{Origin: "播放结束", Subtitle: "播放结束"},
			Next:      model.AutoSub{},
			NextTwo:   model.AutoSub{},
		},
	}
	data, marshalErr := json.Marshal(&_data)
	if marshalErr != nil {
		logger.Err("autoPlay", fmt.Sprintf("[auto] === \n %v \n ===", marshalErr))
		broadcastAutoPlayErr(m, "loop marshal error")
		return
	}
	m.data = data
	WsHub.broadcast <- *m
	return
}

func broadcastAutoPlayEnd(m *message) {
	_data := s2cAutoPlayEnd{
		Head: struct {
			Cmd s2cCmds "json:\"cmd\""
		}{
			Cmd: s2cCmdAutoPlayEnd,
		},
		Body: struct {
			Data interface{} "json:\"data\""
		}{
			Data: "",
		},
	}
	data, marshalErr := json.Marshal(&_data)
	if marshalErr != nil {
		logger.Err("autoPlay", fmt.Sprintf("broadcast auto play end: %v \n", marshalErr))
		return
	}
	m.data = data
	WsHub.broadcast <- *m
	return
}

func broadcastAutoPlayErr(m *message, reason string) {
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
		logger.Err("autoPlay", fmt.Sprintf("broadcast auto play err: %v \n", marshalErr))
		return
	}
	m.data = data
	WsHub.broadcast <- *m
	return
}

func (m *message) handleDeleteAutoSub() error {
	var wsData c2sDeleteAutoSub
	unmarshalErr := json.Unmarshal(m.data, &wsData)
	if unmarshalErr != nil {
		return unmarshalErr
	}
	dbErr := db.DeleteAutoSub(wsData.Body.ListId)
	var _data s2cDeleteAutoSub
	if dbErr != nil {
		_data = s2cDeleteAutoSub{
			Head: struct {
				Cmd s2cCmds "json:\"cmd\""
			}{
				Cmd: s2cCmdDeleteAutoSub,
			},
			Body: struct {
				Status bool "json:\"status\""
				ListId uint "json:\"list_id\""
			}{
				Status: false,
				ListId: wsData.Body.ListId,
			},
		}
	} else {
		_data = s2cDeleteAutoSub{
			Head: struct {
				Cmd s2cCmds "json:\"cmd\""
			}{
				Cmd: s2cCmdDeleteAutoSub,
			},
			Body: struct {
				Status bool "json:\"status\""
				ListId uint "json:\"list_id\""
			}{
				Status: true,
				ListId: wsData.Body.ListId,
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

func (m *message) handleGetAutoPlayStat(rdbCtx context.Context) error {
	var state autoPlayState

	val, rdbErr := db.Rdb.Get(rdbCtx, stat.MakeAutoRdbKey(m.room)).Result()
	if rdbErr == redis.Nil {
		logger.Warn("wsHandler", "redis get auto play stat, the key is expire or undefined")
		(&state).State = stopped
	} else if rdbErr != nil {
		return rdbErr
	} else if val == "" {
		logger.Warn("wsHandler", "redis get auto play stat, the value is empty")
		(&state).State = stopped
	} else {
		unmarshalErr := json.UnmarshalFromString(val, &state)
		if unmarshalErr != nil {
			return unmarshalErr
		}
	}
	_data := s2cGetAutoPlayStat{
		Head: struct {
			Cmd s2cCmds "json:\"cmd\""
		}{
			Cmd: s2cCmdGetAutoPlayStat,
		},
		Body: state,
	}
	data, marshalErr := json.Marshal(&_data)
	if marshalErr != nil {
		return marshalErr
	}
	m.data = data

	WsHub.broadcast <- *m

	return nil
}

func (m *message) handleRecoverPlayStat(ctx context.Context) error {
	var wsData c2sRecoverPlayStat
	unmarshalErr := json.Unmarshal(m.data, &wsData)
	if unmarshalErr != nil {
		return unmarshalErr
	}
	var _data s2cRecoverPlayStat
	rdbErr := db.Rdb.Del(ctx, stat.MakeAutoRdbKey(m.room)).Err()
	dbErr := db.SetRoomListsUnsent(wsData.Body.RoomId)
	if rdbErr != nil || dbErr != nil {
		_data = s2cRecoverPlayStat{
			Head: struct {
				Cmd s2cCmds "json:\"cmd\""
			}{
				Cmd: s2cCmdRecoverPlayStat,
			},
			Body: struct {
				Status bool "json:\"status\""
			}{
				Status: false,
			},
		}
	} else {
		_data = s2cRecoverPlayStat{
			Head: struct {
				Cmd s2cCmds "json:\"cmd\""
			}{
				Cmd: s2cCmdRecoverPlayStat,
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

	WsHub.broadcast <- *m

	return nil
}

func (m *message) handleChangeAutoMemo() error {
	var wsData c2sChangeAutoMemo
	unmarshalErr := json.Unmarshal(m.data, &wsData)
	if unmarshalErr != nil {
		return unmarshalErr
	}
	dbErr := db.ChangeAutoMemo(wsData.Body.ListId, wsData.Body.Memo)

	var _data s2cChangeAutoMemo
	if dbErr != nil {
		_data = s2cChangeAutoMemo{
			Head: struct {
				Cmd s2cCmds "json:\"cmd\""
			}{
				Cmd: s2cCmdChangeAutoMemo,
			},
			Body: struct {
				Status bool   "json:\"status\""
				ListId uint   "json:\"list_id\""
				Memo   string "json:\"memo\""
			}{
				Status: false,
				ListId: wsData.Body.ListId,
				Memo:   wsData.Body.Memo,
			},
		}
	} else {
		_data = s2cChangeAutoMemo{
			Head: struct {
				Cmd s2cCmds "json:\"cmd\""
			}{
				Cmd: s2cCmdChangeAutoMemo,
			},
			Body: struct {
				Status bool   "json:\"status\""
				ListId uint   "json:\"list_id\""
				Memo   string "json:\"memo\""
			}{
				Status: true,
				ListId: wsData.Body.ListId,
				Memo:   wsData.Body.Memo,
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
