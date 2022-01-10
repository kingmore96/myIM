package client

import (
	"context"
	"errors"
	"net"
	"net/url"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	conn   net.Conn
	Closed chan struct{}
	// Recv   chan []byte
}

func (h *Handler) readLoop() error {
	logrus.Info("readloop started")

	for {
		frame, err := ws.ReadFrame(h.conn)
		if err != nil {
			return err
		}
		// if frame.Header.OpCode == ws.OpPong {
		// 	// 重置读取超时时间
		// 	logrus.Info("recv a pong...")
		// 	_ = h.conn.SetReadDeadline(time.Now().Add(h.heartbeat * 3))
		// }

		if frame.Header.OpCode == ws.OpClose {
			return errors.New("remote side close the channel")
		}
		if frame.Header.OpCode == ws.OpText {
			// h.Recv <- frame.Payload
			//直接打印
			logrus.Info("Receive message:", string(frame.Payload))
		}
	}
}

func (h *Handler) SendText(msg string) error {
	logrus.Info("send message :", msg)
	// if err := h.conn.SetWriteDeadline(time.Now().Add(time.Second * 10)); err != nil {
	// 	return err
	// }
	return wsutil.WriteClientText(h.conn, []byte(msg))
}

func Connect(addr string) (*Handler, error) {
	//校验addr
	_, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}

	//连接
	conn, _, _, err := ws.Dial(context.Background(), addr)
	if err != nil {
		return nil, err
	}

	//创建handler
	handler := &Handler{
		conn:   conn,
		Closed: make(chan struct{}, 1),
		// Recv:   make(chan []byte, 1),
	}

	//启动读线程
	go func() {
		err := handler.readLoop()
		if err != nil {
			logrus.Warn("readloop - ", err)
		}
		//通知上层
		handler.Closed <- struct{}{}
	}()

	return handler, nil
}
