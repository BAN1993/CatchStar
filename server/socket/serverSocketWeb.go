package nsocket

import (
	"DrillServerGo/flog"
	"DrillServerGo/public"
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type IOClientWeb struct {
	*IOClientBase
	manager    *ServerSocketWeb
	conn       *websocket.Conn
}

func NewIOClientWeb(s *ServerSocketWeb, c *websocket.Conn) *IOClientWeb {
	return &IOClientWeb{
		IOClientBase: NewIOClientBase(),
		manager:      s,
		conn:         c,
	}
}

func (c *IOClientWeb) GoRead() {
	defer public.CrashCatcher()
	defer func() {
		io := IOClientInterface(c)
		c.manager.ChClose <- &io
		_ = c.conn.Close()
	}()

	c.conn.SetReadLimit(c.maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(c.pongWait))
	c.conn.SetPongHandler(
		func(string) error {
			_= c.conn.SetReadDeadline(time.Now().Add(c.pongWait))
			return nil
		})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		if c.manager.ChRecv != nil {
			io := IOClientInterface(c)
			p := &RecvClientPackage{io: &io, buf:message}
			c.manager.ChRecv <- p
		}
	}
}

func (c *IOClientWeb) GoWrite() {
	defer public.CrashCatcher()

	ticker := time.NewTicker(c.pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case buf, ok := <- c.ChSend:
			_ = c.conn.SetWriteDeadline(time.Now().Add(c.writeWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.BinaryMessage)
			if err != nil {
				return
			}

			_, _ = w.Write(buf)
			if err := w.Close(); err != nil {
				return
			}

		case <- c.ChToClose:
			return

		case <- ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(c.writeWait))
			err := c.conn.WriteMessage(websocket.PingMessage, []byte{})
			if err != nil {
				return
			}
		}
	}
}

type ServerSocketWeb struct {
	*ServerSocketBase
}

func NewServerSocketWeb(cb ServerSocketCallback) *ServerSocketWeb {
	return &ServerSocketWeb{
		ServerSocketBase:NewServerSocketBase(cb),
	}
}

func (s *ServerSocketWeb) Init(addr string) bool {
	http.HandleFunc("/",
		func(res http.ResponseWriter, r *http.Request) {
			ServerWebAcceptHandle(s, res, r)
		})

	err := http.ListenAndServe(addr, nil)
	if err != nil {
		flog.GetInstance().Errorln(err)
		return false
	}

	return true
}

func ServerWebAcceptHandle(s *ServerSocketWeb, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		flog.GetInstance().Errorln(err)
		return
	}

	client := NewIOClientWeb(s, conn)
	io := IOClientInterface(client)
	s.ChAccept <- &io

	go client.GoRead()
	go client.GoWrite()
}
