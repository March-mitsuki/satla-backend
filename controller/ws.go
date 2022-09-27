package controller

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		return origin == "http://localhost:3131"
	},
}

var h = hub{
	broadcast:  make(chan message),
	castother:  make(chan message),
	castself:   make(chan message),
	register:   make(chan subscription),
	unregister: make(chan subscription),
	rooms:      make(map[string]map[*connection]bool),
}

func (h *hub) run() {
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

func (s subscription) readPump() {
	c := s.conn
	defer func() {
		h.unregister <- s
		c.ws.Close()
	}()

	for {
		_, msg, err := c.ws.ReadMessage()
		if err != nil {
			return
		}
		m := message{msg, s.room, s.conn}
		var unmarshal subtitleData
		json.Unmarshal(msg, &unmarshal)
		fmt.Println("-----unmarshal------")
		fmt.Println(unmarshal.Body)
		cmd := unmarshal.Head.Cmd
		switch cmd {
		case "addSubtitle":
			h.broadcast <- m
		case "addUser":
			h.castother <- m
		}
	}
}

func (c *connection) write(mt int, payload []byte) error {
	return c.ws.WriteMessage(mt, payload)
}

func (s subscription) writePump() {
	c := s.conn
	defer func() {
		c.ws.Close()
	}()

	for {
		msg, ok := <-c.send
		if !ok {
			c.write(websocket.CloseMessage, []byte("<-c.send err"))
			return
		}
		fmt.Println("send once")
		if err := c.write(websocket.TextMessage, msg); err != nil {
			return
		}
	}
}

func WsController(c *gin.Context, roomid string) {
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer ws.Close()
	defer fmt.Println("ws closed")

	conn := &connection{send: make(chan []byte, 256), ws: ws}
	sub := subscription{conn: conn, room: roomid}
	h.register <- sub

	go sub.writePump()
	sub.readPump()
}
