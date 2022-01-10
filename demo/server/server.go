package server

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/gobwas/ws"
	"github.com/sirupsen/logrus"
)

type Server struct {
	id      string
	address string
	users   map[string]net.Conn
	mutex   sync.Mutex
}

func NewServer(id, address string) *Server {
	return &Server{
		id:      id,
		address: address,
		users:   make(map[string]net.Conn, 100),
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()
	lg := logrus.WithFields(logrus.Fields{
		"module": "Server",
		"id":     s.id,
		"listen": s.address,
	})
	mux.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		//升级为ws conn
		conn, _, _, err := ws.UpgradeHTTP(r, rw)
		if err != nil {
			lg.Error(err)
			conn.Close()
			return
		}

		//读取userId
		user := r.URL.Query().Get("user")
		if user == "" {
			lg.Error("need user param")
			conn.Close()
			return
		}

		//存储入users数据结构
		old, ok := s.addUser(user, conn)
		if ok {
			//同用户登录互踢
			old.Close()
		}
		lg.Infof("user %s in", user)

		//监听消息
		go func(user string, conn net.Conn) {
			err := s.readLoop(user, conn)
			if err != nil {
				lg.Error(err)
			}
			//关闭连接，清理资源
			conn.Close()
			s.delUser(user)

			lg.Infof("connection of %s closed", user)
		}(user, conn)
	})

	lg.Infoln("Started")
	return http.ListenAndServe(s.address, mux)

}

func (s *Server) addUser(user string, conn net.Conn) (net.Conn, bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	old, ok := s.users[user]
	s.users[user] = conn
	return old, ok
}

func (s *Server) delUser(user string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.users, user)
}

func (s *Server) readLoop(user string, conn net.Conn) error {
	for {
		frame, err := ws.ReadFrame(conn)
		if err != nil {
			return err
		}

		if frame.Header.OpCode == ws.OpClose {
			return errors.New("remote side close the conn")
		}

		if frame.Header.Masked {
			ws.Cipher(frame.Payload, frame.Header.Mask, 0)
		}

		// 接收文本帧内容
		if frame.Header.OpCode == ws.OpText {
			go s.handle(user, string(frame.Payload))
		}
	}
}

//广播操作
func (s *Server) handle(user string, message string) {
	logrus.Infof("recv message %s from %s", message, user)
	s.mutex.Lock()
	defer s.mutex.Unlock()
	broadcast := fmt.Sprintf("%s -- FROM %s", message, user)
	for u, conn := range s.users {
		if u == user { // 不发给自己
			continue
		}
		logrus.Infof("send to %s : %s", u, broadcast)
		// 创建文本帧数据
		f := ws.NewTextFrame([]byte(message))
		err := ws.WriteFrame(conn, f)
		if err != nil {
			logrus.Errorf("write to %s failed, error: %v", user, err)
		}
	}
}

func (s *Server) Shutdown() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for _, conn := range s.users {
		conn.Close()
	}
	logrus.Info("shutdown success")
}
