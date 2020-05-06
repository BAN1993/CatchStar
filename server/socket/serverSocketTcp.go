package nsocket

import (
	"DrillServerGo/flog"
	"DrillServerGo/protocol"
	"DrillServerGo/public"
	"bufio"
	"net"
)

type IOClientTcp struct {
	*IOClientBase
	manager    *ServerSocketTcp
	conn       *net.Conn
}

func NewIOClientTcp(s *ServerSocketTcp, c *net.Conn) *IOClientTcp {
	return &IOClientTcp{
		IOClientBase: NewIOClientBase(),
		manager:      s,
		conn:         c,
	}
}

func (c *IOClientTcp) GoRead() {
	defer public.CrashCatcher()
	defer func() {
		io := IOClientInterface(c)
		c.manager.ChClose <- &io
		_ = (*c.conn).Close()
	}()

	recvBuf := make([]byte, 0)
	reader := bufio.NewReader(*c.conn)
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
			if c.manager.ChRecv != nil {
				io := IOClientInterface(c)
				p := &RecvClientPackage{io: &io, buf: recvBuf[:tmplen]}
				c.manager.ChRecv <- p
			}

			if tmplen == uint16(len(recvBuf)) {
				recvBuf = make([]byte, 0)
			} else {
				recvBuf = recvBuf[tmplen:]
			}
		}
	}
}

func (c *IOClientTcp) GoWrite() {
	defer public.CrashCatcher()
	defer func() {
		_ = (*c.conn).Close()
	}()

	for {
		select {
		case buf := <- c.ChSend:
			if buf == nil { return }

			writer := bufio.NewWriter(*c.conn)
			_, err := writer.Write(buf)
			if err != nil {
				flog.GetInstance().Info(err)
				return
			}
			_ = writer.Flush()

		case <- c.ChToClose:
			return
		}
	}
}

type ServerSocketTcp struct {
	*ServerSocketBase
	listener       net.Listener
}

func NewServerSocketTcp(cb ServerSocketCallback) *ServerSocketTcp {
	return &ServerSocketTcp{
		ServerSocketBase:NewServerSocketBase(cb),
	}
}

// 调用了这个接口会阻塞,所以要放到最后
func (s *ServerSocketTcp) Init(addr string) bool {
	var err error
	s.listener, err = net.Listen("tcp",addr)
	if err != nil {
		flog.GetInstance().Errorln(err)
		return false
	}
	flog.GetInstance().Infof("Listening on %s", addr)

	s.doAccept()
	return true
}

func (s *ServerSocketTcp) doAccept() {
	defer func() {
		flog.GetInstance().Errorf("Listening socket close")
		_ = s.listener.Close()
	}()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			flog.GetInstance().Errorf("Accept error:%s", err)
			continue
		}

		client := NewIOClientTcp(s, &conn)
		io := IOClientInterface(client)
		s.ChAccept <- &io
		go client.GoRead()
		go client.GoWrite()
	}
}
