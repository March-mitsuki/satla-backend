package controllers

import (
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
	c2sCmdHeartBeat          string = "heartBeat"
)

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
	s2cCmdHeartBeat          s2cCmds = "sHeartBeat"
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
	Body struct {
		Subtitle string `json:"subtitle"`
		Origin   string `json:"origin"`
	} `json:"body"`
}

type c2sChangeBilingual struct {
	c2sHead
	Body struct {
		Bilingual bool `json:"bilingual"`
	} `json:"body"`
}

type c2sChangeReversed struct {
	c2sHead
	Body struct {
		Reversed bool `json:"reversed"`
	} `json:"body"`
}

type c2sHeartBeat struct {
	c2sHead
	Body struct {
		Obj string `json:"obj"`
	} `json:"body"`
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
		RoomId         uint   `json:"room_id"`
		NewSubtitleId  uint   `json:"new_subtitle_id"`
		PreSubtitleIdx uint   `json:"pre_subtitle_idx"`
		CheckedBy      string `json:"checked_by"`
	} `json:"body"`
}

type s2cChangeSubtitle struct {
	Head struct {
		Cmd s2cCmds `json:"cmd"`
	} `json:"head"`
	Body struct {
		Status   bool           `json:"status"`
		Subtitle model.Subtitle `json:"subtitle"`
	} `json:"body"`
}

type s2cEditChange struct {
	// start和end只是cmd不一样
	Head struct {
		Cmd s2cCmds `json:"cmd"`
	} `json:"head"`
	Body struct {
		Uname      string `json:"uname"`
		SubtitleId uint   `json:"subtitle_id"`
	} `json:"body"`
}

type s2cAddTranslatedSub struct {
	Head struct {
		Cmd s2cCmds `json:"cmd"`
	} `json:"head"`
	Body struct {
		NewSubtitle model.Subtitle `json:"new_subtitle"`
	} `json:"body"`
}

type s2cDeleteSubtitle struct {
	Head struct {
		Cmd s2cCmds `json:"cmd"`
	} `json:"head"`
	Body struct {
		Status     bool `json:"status"`
		SubtitleId uint `json:"subtitle_id"`
	} `json:"body"`
}

type s2cReorderSub struct {
	Head struct {
		Cmd s2cCmds `json:"cmd"`
	} `json:"head"`
	Body struct {
		OperationUser string `json:"operation_user"`
		Status        bool   `json:"status"`
		DragId        uint   `json:"drag_id"`
		DropId        uint   `json:"drop_id"`
	} `json:"body"`
}

type s2cSendSubtitle struct {
	// 无论哪种发送方式回复都相同
	Head struct {
		Cmd s2cCmds `json:"cmd"`
	} `json:"head"`
	Body struct {
		Status   bool           `json:"status"`
		Subtitle model.Subtitle `json:"subtitle"`
	} `json:"body"`
}

type s2cChangeStyle struct {
	Head struct {
		Cmd s2cCmds `json:"cmd"`
	} `json:"head"`
	Body struct {
		Reversed bool   `json:"reverse"`
		Subtitle string `json:"subtitle"`
		Origin   string `json:"origin"`
	} `json:"body"`
}

type s2cChangeBilingual struct {
	Head struct {
		Cmd s2cCmds `json:"cmd"`
	} `json:"head"`
	Body struct {
		Bilingual bool `json:"bilingual"`
	} `json:"body"`
}

type s2cChangeReversed struct {
	Head struct {
		Cmd s2cCmds `json:"cmd"`
	} `json:"head"`
	Body struct {
		Reversed bool `json:"reversed"`
	} `json:"body"`
}

type s2cHeartBeat struct {
	Head struct {
		Cmd s2cCmds `json:"cmd"`
	} `json:"head"`
	Body struct {
		Data interface{} `json:"data"`
	} `json:"body"`
}
