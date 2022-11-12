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

func (m *message) handlePlayStart(ctx context.Context, ope chan autoOpeData) error {
	var wsData c2sPlayStart
	unmarshalErr := json.Unmarshal(m.data, &wsData)
	if unmarshalErr != nil {
		logger.Err("wsHandler", fmt.Sprintf("unmarshal wsData === \n %v \n ===", unmarshalErr))
		return unmarshalErr
	}
	var autoSubs []model.AutoSub
	result := db.Mdb.Where("list_id = ?", wsData.Body.ListId).Find(&autoSubs)
	if result.Error != nil {
		logger.Err("wsHandler", fmt.Sprintf("[auto] === \n %v \n ===", result.Error))
		broadcastAutoPlayErr(m, "start err")
		return result.Error
	}
	go autoPlayStart(ctx, m, &autoSubs, ope)
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
func autoPlayStartOld(m *message, ctx context.Context, autoSubs []model.AutoSub) {
	defer func() {
		// 结束播放时保证最后一次发送一定是空行
		_lastSend := makeAutoChangeSub(model.AutoSub{})
		data, marshalErr := json.Marshal(&_lastSend)
		if marshalErr != nil {
			logger.Err("autoPlayOld", fmt.Sprintf("[auto] === \n %v \n ===", marshalErr))
			broadcastAutoPlayErr(m, "loop marshal error")
			return
		}
		m.data = data
		WsHub.broadcast <- *m
		logger.Info("autoPlayOld", "Auto Play Finish")
	}()

	logger.Info("autoPlayOld", "auto play start")
	for _, v := range autoSubs {
		_data := makeAutoChangeSub(v)
		data, marshalErr := json.Marshal(&_data)
		if marshalErr != nil {
			logger.Err("autoPlayOld", fmt.Sprintf("[auto] === \n %v \n ===", marshalErr))
			broadcastAutoPlayErr(m, "loop marshal error")
			return
		}
		m.data = data
		d, dErr := time.ParseDuration(fmt.Sprintf("%vs", v.Duration))
		if dErr != nil {
			logger.Err("autoPlayOld", fmt.Sprintf("[auto] === \n %v \n ===", dErr))
			broadcastAutoPlayErr(m, "loop parse duration error")
			return
		}
		WsHub.broadcast <- *m
		var wg sync.WaitGroup
		timer := time.NewTimer(d)
		wg.Add(1)
		go cancelableSleepOld(ctx, timer, &wg)
		wg.Wait()
	}
	return
}

func cancelableSleepOld(ctx context.Context, t *time.Timer, wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
		logger.Info("autoPlayOld", "Sleep done")
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
		go broadcastSendBlank(m)
		go broadcastPreviewEnd(m)
		allAutoCtxs.delCtx(m.room, (*autoSubs)[0].ListId)
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
	var _data s2cAutoPreviewChange
	if nowNum-1 < 0 {
		_data = s2cAutoPreviewChange{
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
				Main:      (*autoSubs)[nowNum],
				Next:      (*autoSubs)[nowNum+1],
				NextTwo:   (*autoSubs)[nowNum+2],
			},
		}
	} else if nowNum-2 < 0 {
		_data = s2cAutoPreviewChange{
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
				Behind:    (*autoSubs)[nowNum-1],
				Main:      (*autoSubs)[nowNum],
				Next:      (*autoSubs)[nowNum+1],
				NextTwo:   (*autoSubs)[nowNum+2],
			},
		}
	} else if nowNum+1 > len(*autoSubs)-1 {
		_data = s2cAutoPreviewChange{
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
				BehindTwo: (*autoSubs)[nowNum-2],
				Behind:    (*autoSubs)[nowNum-1],
				Main:      (*autoSubs)[nowNum],
				Next:      model.AutoSub{},
				NextTwo:   model.AutoSub{},
			},
		}
	} else if nowNum+2 > len(*autoSubs)-1 {
		_data = s2cAutoPreviewChange{
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
				BehindTwo: (*autoSubs)[nowNum-2],
				Behind:    (*autoSubs)[nowNum-1],
				Main:      (*autoSubs)[nowNum],
				Next:      (*autoSubs)[nowNum+1],
				NextTwo:   model.AutoSub{},
			},
		}
	} else {
		_data = s2cAutoPreviewChange{
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
				BehindTwo: (*autoSubs)[nowNum-2],
				Behind:    (*autoSubs)[nowNum-1],
				Main:      (*autoSubs)[nowNum],
				Next:      (*autoSubs)[nowNum+1],
				NextTwo:   (*autoSubs)[nowNum+2],
			},
		}
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
		logger.Err("autoPlay", fmt.Sprintf("make auto play err: %v \n", marshalErr))
		return nil
	}
	m.data = data
	WsHub.broadcast <- *m
	return m
}
