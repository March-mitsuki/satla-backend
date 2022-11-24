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
	var wsData c2sPlayOperation
	unmarshalErr := json.Unmarshal(m.data, &wsData)
	if unmarshalErr != nil {
		logger.Err("wsHandler", fmt.Sprintf("unmarshal wsData === \n %v \n ===", unmarshalErr))
		return unmarshalErr
	}
	_data := s2cAutoPlayOpeRes{
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
	go broadcastAutoChangeSub(m, &subtitle)
	go broadcastAutoPreviewChange(m, autoSubs, 0)

	setAutoStatToRedis(ca, m, autoSubs, playing)

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
		// 结束有3个逻辑, 一个是这里的自动结束
		// 还有一个在cancelableSleep的ctx.Done()里
		// 还有一个在转手动挡之后的manualAutoSend里
		// 自动挡因为是先sleep 结束之后才会call最后一个loop
		// 所以不用给sleep善后
		broadcastCommonStopPlay(ca, m, autoSubs)
		return
	}
	subtitle := (*autoSubs)[ca.adder(1)]
	d, dErr := time.ParseDuration(fmt.Sprintf("%vs", subtitle.Duration))
	if dErr != nil {
		logger.Err("autoPlay", fmt.Sprintf("[auto] === \n %v \n ===", dErr))
		broadcastAutoPlayErr(m, "loop parse duration error")
		return
	}
	go broadcastAutoChangeSub(m, &subtitle)
	go broadcastAutoPreviewChange(m, autoSubs, ca.adder(0))

	setAutoStatToRedis(ca, m, autoSubs, playing)

	timer := time.NewTimer(d)
	wg.Add(1)
	cancelableSleep(ctx, timer, wg, ca, m, autoSubs, ope)
	wg.Wait()

	return
}

func manualAutoSend(
	ctx context.Context,
	ca calcData,
	m *message,
	autoSubs *[]model.AutoSub,
	wg *sync.WaitGroup,
	ope chan autoOpeData,
) {
	logger.Info("autoPlay", "manual send called")
	if ca.adder(0) >= len(*autoSubs)-1 {
		// 手动挡因为sleep的循环一直在监听
		// 所以停止的时候要一起停止sleep的loop
		ctx.Done()
		broadcastCommonStopPlay(ca, m, autoSubs)
		return
	}
	subtitle := (*autoSubs)[ca.adder(1)]
	go broadcastAutoChangeSub(m, &subtitle)
	go broadcastAutoPreviewChange(m, autoSubs, ca.adder(0))

	setAutoStatToRedis(ca, m, autoSubs, playing)

	return
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
	return
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
	return
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
		logger.Nomal("autoPlay", "loop sleep done")
	}()
LOOP:
	for {
		select {
		case <-ctx.Done():
			t.Stop()
			broadcastCommonStopPlay(ca, m, autoSubs)
			return
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
				// pause会暂时接管所有自动播放的接口
				// 并在再次开始或收到停止信号之后break自己的循环
				// 或者收到转手动的信号之后转交手动接管
				logger.Info("wsHandlerAuto", "pause now")
				t.Stop()
				broadcastAutoPlayPause(m, (*autoSubs)[0].ListId)

				setAutoStatToRedis(ca, m, autoSubs, paused)

				for {
					select {
					case <-ctx.Done():
						broadcastCommonStopPlay(ca, m, autoSubs)
						return
					case o := <-ope:
						if o.opeType == restart {
							go autoPlayLoop(ctx, ca, m, autoSubs, wg, ope)
							broadcastAutoPlayRestart(m, (*autoSubs)[0].ListId)
							break
						} else if o.opeType == toManual {
							// 提供接口能够从pause状态转手动
							logger.Info("wsHandlerAuto", "to manual now")
							t.Stop()
							broadcastAtoM(m, (*autoSubs)[0].ListId)

							setAutoStatToRedis(ca, m, autoSubs, manually)

							for {
								select {
								case <-ctx.Done():
									logger.Nomal("wsHandlerAuto", "manual send loop done")
									broadcastCommonStopPlay(ca, m, autoSubs)
									return
								case o := <-ope:
									switch o.opeType {
									case toAuto:
										go autoPlayLoop(ctx, ca, m, autoSubs, wg, ope)
										// manutl to auto 复用 restart 的接口
										broadcastAutoPlayRestart(m, (*autoSubs)[0].ListId)
										// broadcastMtoA(m, (*autoSubs)[0].ListId)
										return
									case foward:
										go manualAutoSend(ctx, ca, m, autoSubs, wg, ope)

									case fowardTwice:
										ca.adder(1)
										go manualAutoSend(ctx, ca, m, autoSubs, wg, ope)

									case rewind:
										ca.suber(2)
										go manualAutoSend(ctx, ca, m, autoSubs, wg, ope)

									case rewindTwice:
										ca.suber(3)
										go manualAutoSend(ctx, ca, m, autoSubs, wg, ope)
									}
								}
							}
						}
					}
				}
			case toManual:
				// 手动播放会接管所有自动播放的接口并改为手动播放
				// 直到播放结束或者手动按下结束按钮
				// 或者直到转回自动挡之后重新开始自动播放并return整个函数
				logger.Info("wsHandlerAuto", "to manual now")
				t.Stop()
				broadcastAtoM(m, (*autoSubs)[0].ListId)

				setAutoStatToRedis(ca, m, autoSubs, manually)

				for {
					select {
					case <-ctx.Done():
						logger.Nomal("wsHandlerAuto", "manual send loop done")
						broadcastCommonStopPlay(ca, m, autoSubs)
						return
					case o := <-ope:
						switch o.opeType {
						case toAuto:
							go autoPlayLoop(ctx, ca, m, autoSubs, wg, ope)
							// manutl to auto 复用 restart 的接口
							broadcastAutoPlayRestart(m, (*autoSubs)[0].ListId)
							// broadcastMtoA(m, (*autoSubs)[0].ListId)
							return
						case foward:
							go manualAutoSend(ctx, ca, m, autoSubs, wg, ope)

						case fowardTwice:
							ca.adder(1)
							go manualAutoSend(ctx, ca, m, autoSubs, wg, ope)

						case rewind:
							ca.suber(2)
							go manualAutoSend(ctx, ca, m, autoSubs, wg, ope)

						case rewindTwice:
							ca.suber(3)
							go manualAutoSend(ctx, ca, m, autoSubs, wg, ope)
						}
					}
				}
			}
			break LOOP
		case <-t.C:
			go autoPlayLoop(ctx, ca, m, autoSubs, wg, ope)
			break LOOP
		}
	}
	return
}

// 发送空白 + 更新preview + 更新operation + 删除ctxRoom + 更新autoStat
func broadcastCommonStopPlay(
	ca calcData,
	m *message,
	autoSubs *[]model.AutoSub,
) {
	go broadcastSendBlank(m)
	go broadcastPreviewEnd(m)
	go broadcastAutoPlayEnd(m)
	allAutoCtxs.delRoom(m.room)
	setAutoStatToRedis(ca, m, autoSubs, stopped)
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

func broadcastAutoPlayPause(m *message, listId uint) {
	_data := s2cAutoPlayOpeRes{}
	_data.Head.Cmd = s2cCmdAutoPlayPause
	_data.Body.ListId = listId

	data, marshalErr := json.Marshal(&_data)
	if marshalErr != nil {
		logger.Err("autoPlay", fmt.Sprintf("broadcast auto play end: %v \n", marshalErr))
		return
	}
	m.data = data
	WsHub.broadcast <- *m
	return
}

func broadcastAutoPlayRestart(m *message, listId uint) {
	_data := s2cAutoPlayOpeRes{}
	_data.Head.Cmd = s2cCmdAutoPlayRestart
	_data.Body.ListId = listId

	data, marshalErr := json.Marshal(&_data)
	if marshalErr != nil {
		logger.Err("autoPlay", fmt.Sprintf("broadcast auto play end: %v \n", marshalErr))
		return
	}
	m.data = data
	WsHub.broadcast <- *m
	return
}

// broadcast Auto to Manual
func broadcastAtoM(m *message, listId uint) {
	_data := s2cAutoPlayOpeRes{}
	_data.Head.Cmd = s2cCmdAutoToManual
	_data.Body.ListId = listId

	data, marshalErr := json.Marshal(&_data)
	if marshalErr != nil {
		logger.Err("autoPlay", fmt.Sprintf("broadcast auto play end: %v \n", marshalErr))
		return
	}
	m.data = data
	WsHub.broadcast <- *m
	return
}

// broadcast Manual to Auto
func broadcastMtoA(m *message, listId uint) {
	_data := s2cAutoPlayOpeRes{}
	_data.Head.Cmd = s2cCmdManualToAuto
	_data.Body.ListId = listId

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

// 检测当前房间状态, 如果当前房间已经是手动状态那么则什么都不做
// func (m *message) handleAtoM() {
// 	rdbCtx := context.Background()
// 	val, rErr := rdbtools.GetValueAndCheck(db.Rdb, rdbCtx, stat.MakeAutoRdbKey(m.room))
// 	if rErr == rdbtools.Empty {

// 	}
// }

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

// will set autoSub[ca.adder(0)] and given stat to redis
func setAutoStatToRedis(
	ca calcData,
	m *message,
	autoSubs *[]model.AutoSub,
	state playState,
) {
	c := context.Background()
	rdbKey := stat.MakeAutoRdbKey(m.room)
	rdbValue := autoPlayState{
		Wsroom:  m.room,
		State:   state,
		ListId:  (*autoSubs)[0].ListId,
		NowSub:  (*autoSubs)[ca.adder(0)],
		Preview: makeAutoPreview(autoSubs, ca.adder(0)),
	}
	rdbValueStr, marshalErr := json.MarshalToString(rdbValue)
	if marshalErr != nil {
		logger.Err("json", fmt.Sprintf("auto play stop manually json marshal: %v", marshalErr))
		return
	}
	rdbErr := db.Rdb.Set(c, rdbKey, rdbValueStr, 5*time.Minute).Err()
	if rdbErr != nil {
		logger.Err("rdb", fmt.Sprintf("auto play pause set: %v", rdbErr))
		return
	}
	return
}
