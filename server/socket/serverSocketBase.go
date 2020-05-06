package nsocket

import (
	"DrillServerGo/public"
	"time"
)

/**
 * 服务Socket回调
 */
type ServerSocketCallback interface {
	// 收到新连接
	OnAcceptClientCallback(io *IOClientInterface)
	// 连接关闭
	OnCloseClientCallback(io *IOClientInterface)
	// 收到客户端消息
	// 保证buf为一个完成的包
	OnRecvClientCallback(io *IOClientInterface, buf []byte)
	// 计时器(1s)
	OnTimerCallback()
}

/**
 * 客户端io接口
 */
type IOClientInterface interface {
	// 主动关闭连接,异步
	// Base已经实现
	Close()
	// 关闭发送管道
	// Base已经实现
	CloseSend()
	// 发送数据
	// Base已经实现
	Send(buf []byte)
	// 读协程
	// 需要自己实现
	GoRead()
	// 写协程
	// 需要自己实现
	GoWrite()

	// 设置或读取一些属性
	// 都有Base提供
	SetId(id uint32)
	GetId() uint32
	SetTempId(id uint32)
	GetTempId() uint32
	SetClientType(t uint32)
	GetClientType() uint32
}

/**
 * 服务端socket接口
 */
type ServerSocketInterface interface {
	// 初始化网络模块
	// 需要自己实现
	Init(addr string) bool
	// 主逻辑协程,负责接收和处理各协程的数据,并上发给业务层
	Run()
	// 注册帧计时器(t为间隔)
	RegistFrameTimer(t time.Duration, cb OnFrameTimerCallback)
}

/**
 * 帧计时器回调
 * 需要调用 RegistFrameTimer 注册
 */
type OnFrameTimerCallback func()

/**
 * 通知逻辑收到的数据包
 * 协程交互用
 */
type RecvClientPackage struct {
	io *IOClientInterface
	buf []byte
}

/**
 * 客户端io基类
 * 为了可以分别实现tcp和ws版本,所以抽出了一个基类
 */
type IOClientBase struct {
	IOClientInterface

	id         uint32
	tempId     uint32
	clientType uint32

	ChSend chan []byte
	ChToClose chan uint8

	// 一些配置项
	writeWait time.Duration
	pongWait time.Duration
	pingPeriod time.Duration
	maxMessageSize int64
}

func NewIOClientBase() *IOClientBase {
	return &IOClientBase{
		id:         0,
		clientType: public.SERVER_TYPE_UNKNOWN,
		ChSend:     make(chan []byte),
		ChToClose:  make(chan uint8),

		writeWait:      10 * time.Second,
		pongWait:       60 * time.Second,
		pingPeriod:     (60 * time.Second * 9) / 10,
		maxMessageSize: 512,
	}
}

func (c *IOClientBase) Close() {
	c.ChToClose <- 1
}

func (c *IOClientBase) CloseSend() {
	close(c.ChSend)
}

func (c *IOClientBase) Send(buf []byte) {
	c.ChSend <- buf
}

func (c *IOClientBase) SetId(id uint32) {
	c.id = id
}

func (c *IOClientBase) GetId() uint32 {
	return c.id
}

func (c *IOClientBase) SetTempId(id uint32) {
	c.tempId = id
}

func (c *IOClientBase) GetTempId() uint32 {
	return c.tempId
}

func (c *IOClientBase) SetClientType(t uint32) {
	c.clientType = t
}

func (c *IOClientBase) GetClientType() uint32 {
	return c.clientType
}

/**
 * 服务Socket基类
 * 为了可以分别实现tcp和ws版本,所以抽出了一个基类
 */
type ServerSocketBase struct {
	ServerSocketInterface

	ChAccept	chan *IOClientInterface
	ChClose		chan *IOClientInterface
	ChRecv		chan *RecvClientPackage

	serverCallback ServerSocketCallback

	// 一些统计数据
	acceptCount	uint32
	closeCount	uint32
	recvCount	uint32
	sendCount	uint32

	// 一些配置
	timerWait time.Duration

	// 帧计时器
	ChFrameTimer chan uint8
	frameTimerCallback OnFrameTimerCallback
}

func NewServerSocketBase(cb ServerSocketCallback) *ServerSocketBase {
	return &ServerSocketBase{
		ChAccept:       make(chan *IOClientInterface),
		ChClose:        make(chan *IOClientInterface),
		ChRecv:         make(chan *RecvClientPackage),
		ChFrameTimer:	make(chan uint8),
		serverCallback: cb,
		acceptCount:    0,
		closeCount:     0,
		recvCount:      0,
		sendCount:      0,
		timerWait:time.Second,
	}
}

func (s *ServerSocketBase) Run() {
	defer public.CrashCatcher()

	ticker := time.NewTicker(s.timerWait)
	defer func() {
		ticker.Stop()
	}()

	for {
		select {
		case io := <-s.ChAccept:
			s.serverCallback.OnAcceptClientCallback(io)
			s.acceptCount++

		case io := <-s.ChClose:
			(*io).CloseSend()
			s.serverCallback.OnCloseClientCallback(io)
			s.closeCount++

		case pack := <-s.ChRecv:
			s.serverCallback.OnRecvClientCallback(pack.io, pack.buf)
			s.recvCount++

		case <-ticker.C:
			s.serverCallback.OnTimerCallback()
			//flog.GetInstance().Infof("[As server]Timer:accept=%d,close=%d,recv=%d,Send=%d", s.acceptCount, s.closeCount, s.recvCount, s.sendCount)
			s.acceptCount = 0
			s.closeCount = 0
			s.recvCount = 0
			s.sendCount = 0

		case <-s.ChFrameTimer:
			s.frameTimerCallback()
		}
	}
}

// 注册帧计时器
func (s *ServerSocketBase) RegistFrameTimer(t time.Duration, cb OnFrameTimerCallback) {
	s.frameTimerCallback = cb
	go s.RunFrameTicker(t)
}

func (s *ServerSocketBase) RunFrameTicker(t time.Duration) {
	defer public.CrashCatcher()

	frameticker := time.NewTicker(t)
	defer func () {
		frameticker.Stop()
	}()

	for {
		select {
		case <- frameticker.C:
			s.ChFrameTimer <- 1
		}
	}
}
