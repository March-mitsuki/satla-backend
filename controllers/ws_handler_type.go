package controllers

import "vvvorld/model"

type roomUsers map[string][]string

// c2s -> client to server
// s2c -> server to client
const (
	c2sCmdChangeUser       string = "changeUser"
	c2sCmdGetRoomSubtitles string = "getRoomSubtitles"
	c2sCmdAddSubtitle      string = "addSubtitle"
)
const (
	s2cCmdChangeUser       string = "sChangeUser"
	s2cCmdGetRoomSubtitles string = "sGetRoomSubtitles"
)

// client会在onopen时发送ChangeUser和getAllSubtitle
type c2sChangeUser struct {
	Head struct {
		Cmd string `json:"cmd"`
	} `json:"head"`
	Body struct {
		Uname string `json:"uname"`
	} `json:"body"`
}

type c2sGetRoomSubtitles struct {
	Head struct {
		Cmd string `json:"cmd"`
	} `json:"head"`
	Body struct {
		Roomid string `json:"roomid"`
	} `json:"body"`
}

type c2sSubtitle struct {
	Head struct {
		Cmd string `json:"cmd"`
	} `json:"head"`
	Body struct {
		Data struct {
			InputTime    string `json:"input_time"`
			SendTime     int64  `json:"send_time"`
			ProjectID    int    `json:"project_id"`
			ProjectName  string `json:"project_name"`
			TranslatedBy string `json:"translated_by"`
			CheckedBy    string `json:"checked_by"`
			Subtitle     string `json:"subtitle"`
			Origin       string `json:"origin"`
		} `json:"data"`
	} `json:"body"`
}

type s2cChangeUser struct {
	Head struct {
		Cmd string `json:"cmd"`
	} `json:"head"`
	Body struct {
		Users []string `json:"users"`
	} `json:"body"`
}

type s2cGetRoomSubtitles struct {
	Head struct {
		Cmd string `json:"cmd"`
	} `json:"head"`
	Body struct {
		Subtitles []model.Subtitle `json:"subtitles"`
	} `json:"body"`
}
