package controllers

import (
	"fmt"
)

var WsHub = hub{
	// 现在的hub没有向所有房间发送消息的逻辑
	// 所有的逻辑都是向对应room（房间）中对应的connection（用户）发送信息
	broadcast:  make(chan message),
	castother:  make(chan message),
	castself:   make(chan message),
	register:   make(chan subscription),
	unregister: make(chan subscription),
	rooms:      make(map[string]map[*connection]bool),
}

func (h *hub) Run() {
	for {
		select {
		case s := <-h.register:
			connections := h.rooms[s.room]
			if connections == nil {
				connections = make(map[*connection]bool)
				h.rooms[s.room] = connections
			}
			h.rooms[s.room][s.conn] = true
			fmt.Println("when <-h.register", h.rooms)
		case s := <-h.unregister:
			connections := h.rooms[s.room]
			fmt.Println("when <-h.unregister", h.rooms)
			if connections != nil {
				if _, ok := connections[s.conn]; ok {
					delete(connections, s.conn)
					close(s.conn.send)
					if len(connections) == 0 {
						delete(h.rooms, s.room)
					}
				}
			}
			fmt.Println("when <-h.unregister", h.rooms)
		case m := <-h.broadcast:
			connections := h.rooms[m.room]
			for c := range connections {
				select {
				case c.send <- m.data:
				default:
					close(c.send)
					delete(connections, c)
					if len(connections) == 0 {
						delete(h.rooms, m.room)
					}
				}
			}
		case m := <-h.castother:
			connections := h.rooms[m.room]
			for c := range connections {
				if m.conn != c {
					select {
					case c.send <- m.data:
					default:
						close(c.send)
						delete(connections, c)
						if len(connections) == 0 {
							delete(h.rooms, m.room)
						}
					}
				}
			}
		case m := <-h.castself:
			connections := h.rooms[m.room]
			for c := range connections {
				if m.conn == c {
					select {
					case c.send <- m.data:
					default:
						close(c.send)
						delete(connections, c)
						if len(connections) == 0 {
							delete(h.rooms, m.room)
						}
					}
				}
			}
		}
	}
}
