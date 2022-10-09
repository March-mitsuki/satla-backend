package controllers

import "vvvorld/model"

type roomUsers map[string][]string

// c2s -> client to server
// s2c -> server to client

const (
	c2sCmdChangeUser       string = "changeUser"
	c2sCmdGetRoomSubtitles string = "getRoomSubtitles"
	c2sCmdAddSubtitleUp    string = "addSubtitleUp"
	c2sCmdAddSubtitleDown  string = "addSubtitleDown"
)

type s2cCmds string

const (
	s2cCmdChangeUser       s2cCmds = "sChangeUser"
	s2cCmdGetRoomSubtitles s2cCmds = "sGetRoomSubtitles"
	s2cCmdAddSubtitleUp    s2cCmds = "sAddSubtitleUp"
	s2cCmdAddSubtitleDown  s2cCmds = "sAddSubtitleDown"
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
		Roomid string `json:"roomid"`
	} `json:"body"`
}

type c2sAddSubtitle struct {
	// 无论up还是down接受的body都相同, 只是cmd不同
	c2sHead
	Body struct {
		PreSubtitleId  uint   `json:"pre_subtitle_id"`
		PreSubtitleIdx uint   `json:"pre_subtitle_idx"`
		ProjectId      uint   `json:"project_id"`
		CheckedBy      string `json:"checked_by"`
	}
}

// ------ 以下 s2c ------

type s2cChangeUser struct {
	Head struct {
		Cmd s2cCmds `json:"cmd"`
	} `json:"head"`
	Body struct {
		Users []string `json:"users"`
	} `json:"body"`
}

type s2cGetRoomSubtitles struct {
	Head struct {
		Cmd s2cCmds `json:"cmd"`
	} `json:"head"`
	Body struct {
		Subtitles []model.Subtitle `json:"subtitles"`
		Order     string           `json:"order"`
	} `json:"body"`
}

type s2cAddSubtitle struct {
	// 无论up还是down回复的body都相同, 只是cmd不同
	Head struct {
		Cmd s2cCmds `json:"cmd"`
	} `json:"head"`
	Body struct {
		ProjectId      uint   `json:"project_id"`
		NewSubtitleId  uint   `json:"new_subtitle_id"`
		PreSubtitleIdx uint   `json:"pre_subtitle_idx"`
		CheckedBy      string `json:"checked_by"`
	} `json:"body"`
}
