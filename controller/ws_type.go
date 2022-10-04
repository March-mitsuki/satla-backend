package controller

import "github.com/gorilla/websocket"

type connection struct {
	ws   *websocket.Conn
	send chan []byte
}

type subscription struct {
	conn *connection
	room string
}

type message struct {
	data []byte
	room string
	conn *connection
}

type hub struct {
	rooms      map[string]map[*connection]bool
	register   chan subscription
	unregister chan subscription
	broadcast  chan message
	castother  chan message
	castself   chan message
}

type wsCmd string

const (
	addUserCmd     wsCmd = "addUser"
	addSubtitleCmd wsCmd = "addSubtitle"
)

type wsAddUserData struct {
	Head struct {
		Cmd wsCmd `json:"cmd"`
	} `json:"head"`
	Body struct {
		Data string `json:"data"`
	} `json:"body"`
}

type wsSubtitleData struct {
	Head struct {
		Cmd wsCmd `json:"cmd"`
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
