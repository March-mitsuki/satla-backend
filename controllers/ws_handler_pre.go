package controllers

import "errors"

func (rUser roomUsers) addUser(wsroom, uname string) {
	// 如果存在该房间则直接追加new user, 若不存在则创建一个新房间并追加
	_, ok := rUser[wsroom]
	if !ok {
		rUser[wsroom] = []string{uname}
		return
	}
	rUser[wsroom] = append(rUser[wsroom], uname)
	return
}

func (rUser roomUsers) delUser(wsroom, uname string) error {
	// 删除指定room内的指定user, 若删除后房间内不存在user, 则会连房间一起删除
	// 若不存在该房间或传入用户名为空值则返回一个error
	if uname == "" {
		return errors.New("[delUser] user name is empty")
	}
	roomuser, ok := rUser[wsroom]
	if !ok {
		return errors.New("[delUser] no such room")
	}
	for idx, v := range roomuser {
		if v == uname {
			// 如果房间内不存在该用户则不会触发删除逻辑, 函数遍历后返回nil
			rUser[wsroom] = append(rUser[wsroom][:idx], rUser[wsroom][idx+1:]...)
			break
		}
	}
	if length := len(rUser[wsroom]); length == 0 {
		delete(rUser, wsroom)
	}
	return nil
}

// 以下自动播放的ctx控制

func (ctx autoCtxs) addCtx(wsroom string, d autoCtxData) {
	_, ok := ctx[wsroom]
	if !ok {
		ctx[wsroom] = []autoCtxData{d}
		return
	}
	ctx[wsroom] = append(ctx[wsroom], d)
	return
}

func (ctx autoCtxs) getCurrentCtx(wsroom string, listId uint) (autoCtxData, error) {
	var result autoCtxData
	roomCtx, ok := ctx[wsroom]
	if !ok {
		return autoCtxData{}, errors.New("[getCurrentCtx] no such room")
	}
	for _, v := range roomCtx {
		if v.listId == listId {
			result = v
			break
		}
	}
	return result, nil
}

func (ctx autoCtxs) delCtx(wsroom string, listId uint) error {
	roomCtx, ok := ctx[wsroom]
	if !ok {
		return errors.New("[delCtx] no such room")
	}
	for idx, v := range roomCtx {
		if v.listId == listId {
			close(ctx[wsroom][idx].opeChan)
			ctx[wsroom] = append(ctx[wsroom][:idx], ctx[wsroom][idx+1:]...)
			break
		}
	}
	if length := len(ctx[wsroom]); length == 0 {
		delete(ctx, wsroom)
	}
	return nil
}

// cancel房间内所有的ctx, 并关闭ope chan, 最后删除整个房间
func (ctx autoCtxs) delRoom(wsroom string) {
	roomCtx, ok := ctx[wsroom]
	if !ok {
		return
	}
	for _, v := range roomCtx {
		v.cancel()
		close(v.opeChan)
	}
	delete(ctx, wsroom)
	return
}
