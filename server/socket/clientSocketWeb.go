package nsocket

/**
 * web客户端连接，只支持websocket
 */

import (
	"DrillServerGo/flog"
	"DrillServerGo/protocol"
	"DrillServerGo/public"
	"github.com/gorilla/websocket"
	"net/url"
	"time"
)

type IOServerWeb struct {
	*IOServerBase
	manager *ClientSocketWeb
	conn *websocket.Conn
}

func newIOServerWeb(s *ClientSocketWeb, id uint32, stype uint32,addr string) *IOServerWeb {
	return &IOServerWeb{
		IOServerBase: NewIOServerBase(id, stype, addr),
		manager : s,
	}
}

func (s *IOServerWeb) GoConnect() {
	defer public.CrashCatcher()

	u := url.URL{Scheme:"ws", Host:s.host, Path:""}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		flog.GetInstance().Errorln(err)
		s.ReConnect()
		return
	}

	s.conn = conn
	io := IOServerInterface(s)
	s.manager.ChConnect <- &io

	go s.goWrite()
	go s.goRead()
}

func (s *IOServerWeb) ReConnect() {
	time.Sleep(time.Second*10) // TODO 先这么处理吧
	go s.GoConnect()
}

func (s *IOServerWeb) goRead() {
	defer public.CrashCatcher()
	defer func() {
		io := IOServerInterface(s)
		s.manager.ChClose <- &io
	}()

	recvBuf := make([]byte, 0)

	for {
		_, line, err := s.conn.ReadMessage()
		if err != nil {
			flog.GetInstance().Infof("Read error and close,err=%s", err)
			return
		}
		linelen := len(line)

		recvBuf = append(recvBuf, line[0:linelen]...)

		for {
			if len(recvBuf) < protocol.HeadLen {
				break
			}

			var head protocol.ProtocolHead
			bio := protocol.Biostream{}
			bio.Attach(recvBuf, len(recvBuf))
			head.ReadHead(&bio)

			if uint16(len(recvBuf)) < head.Length {
				break
			}

			tmplen := head.Length
			if s.manager.ChRecv != nil {
				io := IOServerInterface(s)
				p := &RecvServerPackage{io: &io, buf: recvBuf[:tmplen]}
				s.manager.ChRecv <- p
			}

			// 这里发现一个问题
			// 如果tmplen==len(recvBuf),
			// recvBuf = recvBuf[tmplen:] 会导致recvBuf长度变1024
			if tmplen == uint16(len(recvBuf)) {
				recvBuf = make([]byte, 0)
			} else {
				recvBuf = recvBuf[tmplen:]
			}
		}
	}
}

func (s *IOServerWeb) goWrite() {
	defer public.CrashCatcher()
	defer func() {
		_ = s.conn.WriteMessage(websocket.CloseMessage, []byte{})
		_ = s.conn.Close()
	}()

	for {
		select {
		case buf := <- s.ChSend:
			if buf == nil {
				return
			}

			err := s.conn.WriteMessage(websocket.BinaryMessage, buf)
			if err != nil {
				flog.GetInstance().Info(err)
				return
			}
		}
	}
}

type ClientSocketWeb struct {
	*ClientSocketBase
}

func NewClientSocketWeb(cb ClientSocketCallback) *ClientSocketWeb {
	return &ClientSocketWeb{
		ClientSocketBase:NewClientSocketBase(cb),
	}
}

func (s *ClientSocketWeb) Connect(id uint32, stype uint32, addr string) {
	io := newIOServerWeb(s, id, stype, addr)
	go io.GoConnect()
}
