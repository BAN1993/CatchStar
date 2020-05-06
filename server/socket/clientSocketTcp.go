package nsocket

import (
	"DrillServerGo/flog"
	"DrillServerGo/protocol"
	"DrillServerGo/public"
	"bufio"
	"net"
	"time"
)

type IOServerTcp struct {
	*IOServerBase
	manager *ClientSocketTcp
	conn *net.Conn
}

func newIOServerTcp(s *ClientSocketTcp, id uint32, stype uint32,addr string) *IOServerTcp {
	return &IOServerTcp{
		IOServerBase: NewIOServerBase(id, stype, addr),
		manager:      s,
	}
}

func (s *IOServerTcp) GoConnect() {
	defer public.CrashCatcher()

	conn, err := net.Dial("tcp", s.host)
	if err != nil {
		flog.GetInstance().Errorln(err)
		s.ReConnect()
		return
	}

	s.conn = &conn
	io := IOServerInterface(s)
	s.manager.ChConnect <- &io

	go s.goWrite()
	go s.goRead()
}

func (s *IOServerTcp) ReConnect() {
	time.Sleep(time.Second*10) // TODO 先这么处理吧
	go s.GoConnect()
}

func (s *IOServerTcp) goRead() {
	defer public.CrashCatcher()
	defer func() {
		io := IOServerInterface(s)
		s.manager.ChClose <- &io
	}()

	recvBuf := make([]byte, 0)
	reader := bufio.NewReader(*s.conn)
	for {
		line := make([]byte, 1024)
		linelen, err := reader.Read(line)
		if err != nil {
			flog.GetInstance().Errorln(err)
			return
		}

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

func (s *IOServerTcp) goWrite() {
	defer public.CrashCatcher()
	defer func() {
		_ = (*s.conn).Close()
	}()

	for {
		select {
		case buf := <- s.ChSend:
			if buf == nil { return }
			writer := bufio.NewWriter(*s.conn)
			_, err := writer.Write(buf)
			if err != nil {
				flog.GetInstance().Info(err)
				return
			}
			_ = writer.Flush()
		}
	}
}

type ClientSocketTcp struct {
	*ClientSocketBase
}

func NewClientSocketTcp(cb ClientSocketCallback) *ClientSocketTcp {
	return &ClientSocketTcp{
		ClientSocketBase:NewClientSocketBase(cb),
	}
}

func (s *ClientSocketTcp) Connect(id uint32, stype uint32, addr string) {
	io := newIOServerTcp(s, id, stype, addr)
	go io.GoConnect()
}
