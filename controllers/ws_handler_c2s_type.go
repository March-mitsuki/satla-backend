package controllers

import (
	"github.com/March-mitsuki/satla-backend/model"
)

// 定义一个可复用的c2s head方便编写
type c2sHead struct {
	Head struct {
		Cmd string `json:"cmd"`
	} `json:"head"`
}

// client会在onopen时发送ChangeUser和getRoomSubtitles
type c2sChangeUser struct {
	c2sHead
	Body struct {
		Uname string `json:"uname"`
	} `json:"body"`
}

// client会在onopen时发送ChangeUser和getRoomSubtitles
type c2sGetRoomSubtitles struct {
	c2sHead
	Body struct {
		RoomId uint `json:"room_id"`
	} `json:"body"`
}

type c2sAddSubtitle struct {
	// 无论up还是down接受的body都相同, 只是cmd不同
	c2sHead
	Body struct {
		PreSubtitleId  uint   `json:"pre_subtitle_id"`
		PreSubtitleIdx uint   `json:"pre_subtitle_idx"`
		RoomId         uint   `json:"room_id"`
		CheckedBy      string `json:"checked_by"`
	} `json:"body"`
}

type c2sChangeSubtitle struct {
	c2sHead
	Body struct {
		model.Subtitle `json:"subtitle"`
	} `json:"body"`
}

type c2sEditChange struct {
	// start和end只是cmd不同
	c2sHead
	Body struct {
		Uname      string `json:"uname"`
		SubtitleId uint   `json:"subtitle_id"`
	} `json:"body"`
}

type c2sAddTranslatedSub struct {
	c2sHead
	Body struct {
		NewSubtitle model.Subtitle `json:"new_subtitle"`
	} `json:"body"`
}

type c2sDeleteSubtitle struct {
	c2sHead
	Body struct {
		Subtitle model.Subtitle `json:"subtitle"`
	} `json:"body"`
}

type c2sReorderSub struct {
	// front和back只是cmd不同
	c2sHead
	Body struct {
		OperationUser string `json:"operation_user"`
		RoomId        uint   `json:"room_id"`
		DragId        uint   `json:"drag_id"`
		DropId        uint   `json:"drop_id"`
	} `json:"body"`
}

type c2sSendSubtitle struct {
	c2sHead
	Body struct {
		Subtitle model.Subtitle `json:"subtitle"`
	} `json:"body"`
}

type c2sSendSubtitleDirect struct {
	c2sHead
	Body struct {
		Subtitle model.Subtitle `json:"subtitle"`
	} `json:"body"`
}

type c2sChangeStyle struct {
	c2sHead
	Body ChangeStyleBody `json:"body"`
}

type c2sBatchAddSubs struct {
	c2sHead
	Body struct {
		Subtitles []model.Subtitle `json:"subtitles"`
	} `json:"body"`
}

type c2sGetNowRoomStyle struct {
	c2sHead
	Body struct {
		Wsroom string `json:"wsroom"`
	} `json:"body"`
}

type c2sGetNowRoomSub struct {
	c2sHead
	Body struct {
		Wsroom string `json:"wsroom"`
	} `json:"body"`
}

//
// 以下为auto page
//

type c2sGetAutoLists struct {
	c2sHead
	Body struct {
		RoomId uint `json:"room_id"`
	} `json:"body"`
}

type c2sAddAutoSub struct {
	c2sHead
	Body struct {
		AutoSubs []model.AutoSub `json:"auto_subs"`
		Memo     string          `json:"memo"`
	} `json:"body"`
}

// 以下cmd均遵循这个格式, 但使用jsoniter.Get来获取元素, 不解构到struct中
//
// play start
//
// play end
//
// play forward (twice)
//
// play rewind (twice)
//
// play pause
//
// play restart
//
// auto to manual
//
// manual to auto
type c2sPlayOperation struct {
	c2sHead
	Body struct {
		ListId uint `json:"list_id"`
	} `json:"body"`
}

type c2sDeleteAutoSub struct {
	c2sHead
	Body struct {
		ListId uint `json:"list_id"`
	} `json:"body"`
}

type c2sGetAutoPlayStat struct {
	c2sHead
	Body struct {
		Data interface{} `json:"data"`
	} `json:"body"`
}

type c2sRecoverPlayStat struct {
	c2sHead
	Body struct {
		RoomId uint `json:"room_id"`
	} `json:"body"`
}

type c2sChangeAutoMemo struct {
	c2sHead
	Body struct {
		ListId uint   `json:"list_id"`
		Memo   string `json:"memo"`
	} `json:"body"`
}

// heartbeat用于检测当前连接房间的type与连接状态, 会在onopen的时候发送
// 并且每30秒由client发给server一次
type c2sHeartBeat struct {
	c2sHead
	Body struct {
		RoomType RoomType `json:"room_type"`
		RoomId   uint     `json:"room_id"`
	} `json:"body"`
}

type RoomType string

const (
	nomal RoomType = "nomal"
	auto  RoomType = "auto"
)
