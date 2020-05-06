package nsocket

import (
	"DrillServerGo/protocol"
	"DrillServerGo/public"
	"time"
)

/**
 * 客户端Socket回调
 */
type ClientSocketCallback interface {
	// 连接成功
	OnConnectServerCallback(c *IOServerInterface)
	// 连接关闭
	OnCloseServerCallback(c *IOServerInterface)
	// 收到服务端消息
	// 保证buf为一个完整的包
	OnRecvServerCallback(c *IOServerInterface, buf []byte)
}

/**
 * 服务io接口
 * TODO 尽量在Base实现大部分逻辑
 */
type IOServerInterface interface {
	/**
	 * 以下接口需要上层自己实现
	 * 主要是ws和tcp之间有差异化的地方
	 */
	// 连接协程,连接成功后开启读写协程
	GoConnect()
	// 读协程,负责处理粘包拆包,将完成协议上发给主逻辑
	GoRead()
	// 写协程
	GoWrite()

	/**
	 * 以下接口皆由Base实现,直接调用即可
	 */
	// 重连,现在处理比较简单,睡眠n秒后调用连接协程
	ReConnect()
	// 发送数据
	Send(buf []byte)
	// 发送心跳,放在timer里调用即可,内部控制1分钟发一次
	TryHeartBeat()
	// 设置或读取一些属性
	SetId(id uint32)
	GetId() uint32
	SetServerType(t uint32)
	GetServerType() uint32
}

/**
 * 客户端socket接口
 */
type ClientSocketInterface interface {
	// 负责接收和处理各协程的数据,并上发给业务层
	Run()
	// 异步连接接口
	Connect(id uint32, stype uint32, addr string)
}

/**
 * 通知逻辑收到的数据包
 * 协程交互用
 */
type RecvServerPackage struct {
	io *IOServerInterface
	buf []byte
}

/**
 * 客户端io基类
 * 为了可以分别实现tcp和ws版本,所以抽出了一个基类
 */
type IOServerBase struct {
	IOServerInterface

	id         uint32
	serverType uint32
	host string
	ChSend chan []byte
	lastHeartTime int64
}

func NewIOServerBase(id uint32, stype uint32, addr string) *IOServerBase {
	return &IOServerBase{
		id:            id,
		serverType:    stype,
		host:          addr,
		ChSend:        make(chan []byte),
		lastHeartTime: 0,
	}
}

func (s *IOServerBase) TryHeartBeat() {
	// 这个地方由server主线程调用
	now := time.Now().Unix()
	if s.lastHeartTime==0 || now-s.lastHeartTime>=60 {
		var req protocol.HeartBeat
		req.Timestamp = uint32(now)
		sendbuf := make([]byte, 128)
		len := protocol.ProtocolToBuffer(&req, sendbuf)
		s.Send(sendbuf[:len])
		s.lastHeartTime = now
	}
}

func (s *IOServerBase) Send(buf []byte) {
	s.ChSend <- buf
}

func (s *IOServerBase) SetId(id uint32) {
	s.id = id
}

func (s *IOServerBase) GetId() uint32 {
	return s.id
}

func (s *IOServerBase) SetServerType(t uint32) {
	s.serverType = t
}

func (s *IOServerBase) GetServerType() uint32 {
	return s.serverType
}

/**
 * 客户端Socket基类
 * 为了可以分别实现tcp和ws版本,所以抽出了一个基类
 */
type ClientSocketBase struct {
	ClientSocketInterface

	ChConnect	chan *IOServerInterface
	ChClose		chan *IOServerInterface
	ChRecv		chan *RecvServerPackage

	clientCallback ClientSocketCallback
}

func NewClientSocketBase(cb ClientSocketCallback) *ClientSocketBase {
	return &ClientSocketBase{
		ChConnect:	make(chan *IOServerInterface),
		ChClose:	make(chan *IOServerInterface),
		ChRecv:		make(chan *RecvServerPackage),
		clientCallback:	cb,
	}
}

func (s *ClientSocketBase) Run() {
	defer public.CrashCatcher()

	for {
		select {
		case server := <-s.ChConnect:
			s.clientCallback.OnConnectServerCallback(server)

		case server := <-s.ChClose:
			s.clientCallback.OnCloseServerCallback(server)
			(*server).ReConnect()

		case pack := <-s.ChRecv:
			s.clientCallback.OnRecvServerCallback(pack.io, pack.buf)
		}
	}
}
