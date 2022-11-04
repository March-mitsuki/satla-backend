package db

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
