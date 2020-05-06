package main

import (
	"DrillServerGo/flog"
	"DrillServerGo/protocol"
	"DrillServerGo/public"
	"DrillServerGo/socket"
	"fmt"
	"os"
)

var gConfig Configs

type RobotClient struct {
	host         string
	account      string
	password     string
	clientSocket nsocket.ClientSocketInterface

	numid    uint32
	nickname string
}

func NewRobotClient(addr string, acc string, pwd string) *RobotClient {
	client := RobotClient{
		host:     addr,
		account:  acc,
		password: pwd,
	}

	var socket nsocket.ClientSocketInterface
	if gConfig.mode == "ws" {
		flog.GetInstance().Infof("Init Web ClientSocket")
		socket = nsocket.NewClientSocketWeb(&client)
	} else {
		flog.GetInstance().Infof("Init Tcp ClientSocket")
		socket = nsocket.NewClientSocketTcp(&client)
	}
	client.clientSocket = socket
	return &client
}

func (c *RobotClient) OnConnectServerCallback(io *nsocket.IOServerInterface) {
	flog.GetInstance().Infof("connect success,id=%d,type=%d", (*io).GetId(), (*io).GetServerType())

	var req protocol.ReqLogin
	req.Account = c.account
	req.Password = c.password
	sendbuf := make([]byte, 128)
	len := protocol.ProtocolToBuffer(&req, sendbuf)
	(*io).Send(sendbuf[:len])
}

func (c *RobotClient) OnCloseServerCallback(io *nsocket.IOServerInterface) {
	flog.GetInstance().Infof("server closed,id=%d,type=%d", (*io).GetId(), (*io).GetServerType())
}

func (c *RobotClient) OnRecvServerCallback(io *nsocket.IOServerInterface, buf []byte) {
	var head protocol.ProtocolHead
	bio := protocol.Biostream{}
	bio.Attach(buf, len(buf))
	head.ReadHead(&bio)
	flog.GetInstance().Debugf("xyid=%d,len=%d,numid=%d,id=%d,type=%d", head.Xyid, head.Length, head.Numid, (*io).GetId(), (*io).GetServerType())

	switch head.Xyid {
	case protocol.XYID_RES_LOGIN:
		var res protocol.ResLogin
		protocol.BufferToProtocol(buf, &res)
		flog.GetInstance().Infof("ResLogin,flag=%d,numid=%d,nickanme=%s", res.Flag, res.Numid, res.Nickname)

		if res.Flag == 0 {

			c.numid = res.Numidxy
			c.nickname = res.Nickname

			var req protocol.ReqJoinRoom
			req.Numid = c.numid
			req.Nickname = c.nickname
			sendbuf := make([]byte, 128)
			len := protocol.ProtocolToBuffer(&req, sendbuf)
			(*io).Send(sendbuf[:len])
		}

	case protocol.XYID_RES_JOINROOM:
		var res protocol.ResJoinRoom
		protocol.BufferToProtocol(buf, &res)
		flog.GetInstance().Infof("ResJoinRoom,flag=%d,numid=%d", res.Flag, res.Numid)

		if res.Flag == 0 {
			//
		}

	case protocol.XYID_NTF_GAMESTART:
		var ntf protocol.NtfGameStart
		protocol.BufferToProtocol(buf, &ntf)
		flog.GetInstance().Infof("NtfGameStart,numid=%d", ntf.Numid)
		var req protocol.ReqConnect
		req.Numid = c.numid
		sendbuf := make([]byte, 128)
		len := protocol.ProtocolToBuffer(&req, sendbuf)
		(*io).Send(sendbuf[:len])

		// TODO 还没有做控制和游戏界面，先用c++版的客户端
		//关掉屏幕日志，开启控制和游戏界面
		//flog.GetInstance().SetLevel(flog.LevelError) // 没有关闭屏幕输出的好方法，先改日志等级吧

	case protocol.XYID_NTF_GAMEEND:
		var ntf protocol.NtfGameEnd
		protocol.BufferToProtocol(buf, &ntf)
		flog.GetInstance().Infof("NtfGameEnd,numid=%d", ntf.Numid)
		//flog.GetInstance().SetLevel(flog.LevelDebug)
	default:
		break
	}
}

func (c* RobotClient)connect() {
	c.clientSocket.Connect(1,public.SERVER_TYPE_GATEWAY, c.host)
}

func (c* RobotClient) run() {
	c.clientSocket.Run()
}

func main() {
	err := gConfig.Load("client_config.ini")
	if err != nil {
		fmt.Println("Load config error:", err)
		return
	}

	err = flog.InitLog(gConfig.logfilename, gConfig.loglevel)
	if err != nil {
		fmt.Println(os.Stderr, "InitLog error:", err)
		return
	}

	var acc string
	var pwd string
	fmt.Println("Input account ")
	fmt.Scanln(&acc)
	fmt.Println("Input password: ")
	fmt.Scanln(&pwd)

	robot := NewRobotClient(gConfig.serverhost, acc, pwd)
	robot.connect()
	robot.run()
}
