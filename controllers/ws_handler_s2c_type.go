package controllers

import "github.com/March-mitsuki/satla-backend/model"

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

// 以下为auto page

type s2cGetAutoLists struct {
	Head struct {
		Cmd s2cCmds `json:"cmd"`
	} `json:"head"`
	Body struct {
		Status    bool             `json:"status"`
		AutoLists []model.AutoList `json:"auto_lists"`
	} `json:"body"`
}

type s2cAddAutoSub struct {
	Head struct {
		Cmd s2cCmds `json:"cmd"`
	} `json:"head"`
	Body struct {
		Status  bool           `json:"status"`
		NewList model.AutoList `json:"new_list"`
	} `json:"body"`
}

type s2cAutoPlayErr struct {
	Head struct {
		Cmd s2cCmds `json:"cmd"`
	} `json:"head"`
	Body struct {
		Msg string `json:"msg"`
	} `json:"body"`
}

type s2cAutoChangeSub struct {
	Head struct {
		Cmd s2cCmds `json:"cmd"`
	} `json:"head"`
	Body struct {
		Subtitle model.AutoSub `json:"subtitle"`
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
