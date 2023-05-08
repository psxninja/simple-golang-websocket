package main

import (
	"fmt"
	"gotest/wsserver"
	"net/http"

	"golang.org/x/net/websocket"
)

var users = map[string]*wsserver.Session{}

func main() {
	ws := wsserver.New()
	http.Handle("/ws", websocket.Handler(ws.HandleRequest))
	http.Handle("/", http.FileServer(http.Dir("./public")))

	ws.HandleConnect(func(s *wsserver.Session) {
		resquest := s.Conn.Request()
		domain := resquest.Host

		users[s.Uuid] = s

		fmt.Println(domain)
	})
	ws.HandleMessage(func(s *wsserver.Session, msg []byte) {
		fmt.Println(string(msg))
		ws.BroadcastOthers(s, msg)
	})
	ws.HandleDisconnect(func(s *wsserver.Session) {
		delete(users, s.Uuid)
		fmt.Println("client disconnected")
	})

	http.ListenAndServe(":9009", nil)
}
