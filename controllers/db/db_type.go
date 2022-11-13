package db

import "github.com/March-mitsuki/satla-backend/model"

type ArgAddSubtitle struct {
	RoomId        uint
	PreSubtitleId uint
	CheckedBy     string
}

type ArgChangeSubtitle struct {
	ID        uint
	CheckedBy string
	Subtitle  string
	Origin    string
}

type ArgChangeUserPassword struct {
	ID      uint
	OldPass string
	NewPass string
}

type ArgAddAutoSub struct {
	AutoSubs []model.AutoSub
	Memo     string
}
