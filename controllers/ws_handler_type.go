package controllers

import (
	"context"

	"github.com/March-mitsuki/satla-backend/model"
)

type roomUsers map[string][]string

type SubtitleFromClient struct {
	ID           uint        `json:"id"`
	CreatedAt    interface{} `json:"-"`
	UpdatedAt    interface{} `json:"-"`
	DeletedAt    interface{} `json:"-"`
	InputTime    string      `json:"input_time"`
	SendTime     interface{} `json:"send_time"`
	ProjectID    uint        `json:"project_id"`
	TranslatedBy string      `json:"translated_by"`
	CheckedBy    string      `json:"checked_by"`
	Subtitle     string      `json:"subtitle"`
	Origin       string      `json:"origin"`
}

// 自动播放时使用的ctx组合
type autoCtxData struct {
	ctx     context.Context
	cancel  context.CancelFunc
	listId  uint
	opeChan chan autoOpeData
}
type autoCtxs map[string][]autoCtxData
type autoOpeData struct {
	opeType opeCmd
}
type opeCmd uint

const (
	foward opeCmd = iota
	fowardTwice
	rewind
	rewindTwice
	pause
	restart
)

type autoPreview struct {
	BehindTwo model.AutoSub `json:"behind_two"`
	Behind    model.AutoSub `json:"behind"`
	Main      model.AutoSub `json:"main"`
	Next      model.AutoSub `json:"next"`
	NextTwo   model.AutoSub `json:"next_two"`
}

// redis储存当前房间状态

type playState int

const (
	stopped playState = iota
	playing
	paused
)

type autoPlayState struct {
	Wsroom  string        `json:"wsroom"`
	State   playState     `json:"state"`
	ListId  uint          `json:"list_id"`
	NowSub  model.AutoSub `json:"now_sub"`
	Preview autoPreview   `json:"preview"`
}

// c2s -> client to server
// s2c -> server to client

const (
	c2sCmdChangeUser         string = "changeUser"
	c2sCmdGetRoomSubtitles   string = "getRoomSubtitles"
	c2sCmdAddSubtitleUp      string = "addSubtitleUp"
	c2sCmdAddSubtitleDown    string = "addSubtitleDown"
	c2sCmdChangeSubtitle     string = "changeSubtitle"
	c2sCmdEditStart          string = "editStart"
	c2sCmdEditEnd            string = "editEnd"
	c2sCmdAddTranslatedSub   string = "addTransSub"
	c2sCmdDeleteSubtitle     string = "deleteSubtitle"
	c2sCmdReorderSubFront    string = "reorderSubFront" // 从前往后拖
	c2sCmdReorderSubBack     string = "reorderSubBack"  // 从后往前拖
	c2sCmdSendSubtitle       string = "sendSubtitle"
	c2sCmdSendSubtitleDirect string = "sendSubtitleDirect"
	c2sCmdChangeStyle        string = "changeStyle"
	c2sCmdChangeBilingual    string = "changeBilingual"
	c2sCmdChangeReversed     string = "changeReversed"
)
const (
	c2sCmdGetAutoLists     string = "getRoomAutoLists"
	c2sCmdAddAutoSub       string = "addAutoSub"
	c2sCmdPlayStart        string = "playStart"
	c2sCmdPlayEnd          string = "playEnd"
	c2sCmdPlayForward      string = "playForward"
	c2sCmdPlayForwardTwice string = "playForwardTwice"
	c2sCmdPlayRewind       string = "playRewind"
	c2sCmdPlayRewindTwice  string = "playRewindTwice"
	c2sCmdPlayPause        string = "playPause"
	c2sCmdPlayRestart      string = "playRestart"
	c2sCmdPlaySendBlank    string = "playSendBlank"
	c2sCmdDeleteAutoSub    string = "deleteAutoSub"
	c2sCmdGetAutoPlayStat  string = "getAutoPlayStat"
	c2sCmdRecoverPlayStat  string = "recoverAutoPlayStat"
)
const c2sCmdHeartBeat string = "heartBeat"

type s2cCmds string

const (
	s2cCmdChangeUser         s2cCmds = "sChangeUser"
	s2cCmdGetRoomSubtitles   s2cCmds = "sGetRoomSubtitles"
	s2cCmdAddSubtitleUp      s2cCmds = "sAddSubtitleUp"
	s2cCmdAddSubtitleDown    s2cCmds = "sAddSubtitleDown"
	s2cCmdChangeSubtitle     s2cCmds = "sChangeSubtitle"
	s2cCmdEditStart          s2cCmds = "sEditStart"
	s2cCmdEditEnd            s2cCmds = "sEditEnd"
	s2cCmdAddTranslatedSub   s2cCmds = "sAddTransSub"
	s2cCmdDeleteSubtitle     s2cCmds = "sDeleteSubtitle"
	s2cCmdReorderSubFront    s2cCmds = "sReorderSubFront"
	s2cCmdReorderSubBack     s2cCmds = "sReorderSubBack"
	s2cCmdSendSubtitle       s2cCmds = "sSendSubtitle"
	s2cCmdSendSubtitleDirect s2cCmds = "sSendSubtitleDirect"
	s2cCmdChangeStyle        s2cCmds = "sChangeStyle"
	s2cCmdChangeBilingual    s2cCmds = "sChangeBilingual"
	s2cCmdChangeReversed     s2cCmds = "sChangeReversed"
)
const (
	s2cCmdGetAutoLists      s2cCmds = "sGetRoomAutoLists"
	s2cCmdAddAutoSub        s2cCmds = "sAddAutoSub"
	s2cCmdAutoPlayErr       s2cCmds = "autoPlayErr"
	s2cCmdAutoChangeSub     s2cCmds = "autoChangeSub"
	s2cCmdAutoPreviewChange s2cCmds = "autoPreviewChange"
	s2cCmdAutoPlayEnd       s2cCmds = "autoPlayEnd"
	s2cCmdDeleteAutoSub     s2cCmds = "sDeleteAutoSub"
	s2cCmdGetAutoPlayStat   s2cCmds = "sGetAutoPlayStat"
	s2cCmdRecoverPlayStat   s2cCmds = "sRecoverAutoPlayStat"
)
const s2cCmdHeartBeat s2cCmds = "sHeartBeat"
