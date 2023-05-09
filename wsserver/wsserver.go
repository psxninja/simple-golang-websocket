package wsserver

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"golang.org/x/net/websocket"
)

type Server struct {
	rwmutex           sync.RWMutex
	conns             map[*Session]bool
	connectHandler    func(*Session)
	messageHandler    func(*Session, []byte)
	disconnectHandler func(*Session)
	errorHandler      func(*Session, error)
}

type Session struct {
	Uuid string
	Conn *websocket.Conn
}

func (s *Session) Write(msg []byte) (n int, err error) {
	nn, errr := s.Conn.Write(msg)
	return nn, errr
}

func (s *Session) CloseWithStatus(code int) {
	s.Conn.WriteClose(code)
}

func New() *Server {
	return &Server{
		conns:             map[*Session]bool{},
		connectHandler:    func(*Session) {},
		messageHandler:    func(*Session, []byte) {},
		disconnectHandler: func(*Session) {},
		errorHandler:      func(*Session, error) {},
	}
}

func (s *Server) HandleConnect(fn func(*Session)) {
	s.connectHandler = fn
}

func (s *Server) HandleMessage(fn func(*Session, []byte)) {
	s.messageHandler = fn
}

func (s *Server) HandleDisconnect(fn func(*Session)) {
	s.disconnectHandler = fn
}

func (s *Server) HandleError(fn func(*Session, error)) {
	s.errorHandler = fn
}

func (s *Server) HandleRequest(ws *websocket.Conn) {
	s.rwmutex.Lock()
	ws.MaxPayloadBytes = 1024 * 1024
	uuid := s.guuid(15)
	session := &Session{
		Uuid: uuid,
		Conn: ws,
	}
	s.conns[session] = true
	s.connectHandler(session)
	s.rwmutex.Unlock()
	s.readLoop(session)
}

func (s *Server) Close(ss *Session) {
	s.rwmutex.Lock()
	defer s.rwmutex.Unlock()
	s.disconnectHandler(ss)
	delete(s.conns, ss)
}

func (s *Server) readLoop(ss *Session) {
	defer s.Close(ss)
	for {
		var msg []byte
		err := websocket.Message.Receive(ss.Conn, &msg)
		if err != nil {
			s.errorHandler(ss, err)
			s.Close(ss)
			break
		}
		s.messageHandler(ss, msg)
	}
}

func (s *Server) BroadcastOthers(ss *Session, b []byte) {
	for c := range s.conns {
		if c.Uuid == ss.Uuid {
			continue
		}
		go func(ws *Session) {
			err := websocket.Message.Send(ws.Conn, b)
			if err != nil {
				fmt.Println("write error")
				return
			}
		}(c)
	}
}

func (s *Server) guuid(length int) string {
	rand.Seed(time.Now().UnixNano())
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	chatsetLen := len(charset)
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(chatsetLen)]
	}
	return string(b)
}
