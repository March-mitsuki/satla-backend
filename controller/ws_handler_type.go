package controller

type roomUsers map[string][]string

// c2s -> client to server
// s2c -> server to client
const (
	c2sCmdAddUser     string = "addUser"
	c2sCmdAddSubtitle string = "addSubtitle"
)
const (
	s2cCmdAddUser string = "sAddUser"
)

type c2sAddUser struct {
	Head struct {
		Cmd string `json:"cmd"`
	} `json:"head"`
	Body struct {
		Uname string `json:"uname"`
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

type s2cAddUser struct {
	Head struct {
		Cmd string `json:"cmd"`
	} `json:"head"`
	Body struct {
		Users []string `json:"users"`
	} `json:"body"`
}
