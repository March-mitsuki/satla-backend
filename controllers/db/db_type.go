package db

type ArgAddSubtitle struct {
	ProjectId     uint
	PreSubtitleId uint
	CheckedBy     string
}

type ArgChangeSubtitle struct {
	ID        uint
	CheckedBy string
	Subtitle  string
	Origin    string
}
