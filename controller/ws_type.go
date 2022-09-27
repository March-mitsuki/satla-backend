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

type subtitleData struct {
	Head struct {
		Cmd string `json:"cmd"`
	} `json:"head"`
	Body struct {
		Data struct {
			InputTime    string      `json:"input_time"`
			SendTime     int64       `json:"send_time"`
			ProjectID    int         `json:"project_id"`
			ProjectName  string      `json:"project_name"`
			TranslatedBy string      `json:"translated_by"`
			CheckedBy    interface{} `json:"checked_by"`
			Subtitle     string      `json:"subtitle"`
			Origin       string      `json:"origin"`
		} `json:"data"`
	} `json:"body"`
}
