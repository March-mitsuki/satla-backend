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
