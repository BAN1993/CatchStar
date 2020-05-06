package main

/**
 * 发现一个bug:opt多的时候,重连会有错位的情况,后面有时间改
 */

import (
	"DrillServerGo/flog"
	"DrillServerGo/protocol"
	"DrillServerGo/public"
	"DrillServerGo/socket"
	"container/list"
	"fmt"
	"math/rand"
	"time"
)

type Server struct {
	serverConfig  Configs
	playerManager *PlayerManager

	dbsList      map[uint32]*nsocket.IOServerInterface
	gwList       map[uint32]*nsocket.IOClientInterface
	clientSocket *nsocket.ClientSocketTcp
	serverSocket *nsocket.ServerSocketTcp

	gameIdIndex uint64
	matchList   list.List // <Player>
	gameList    map[uint64]*Game
}

func (server *Server) init() error {
	rand.Seed(time.Now().UnixNano())

	err := server.serverConfig.Load("gs_config.ini")
	if err != nil {
		fmt.Println("Load config error:", err)
		return err
	}

	err = flog.InitLog(server.serverConfig.logfilename, server.serverConfig.loglevel)
	if err != nil {
		fmt.Println("InitLog error:", err)
		return err
	}

	server.dbsList = make(map[uint32]*nsocket.IOServerInterface)
	server.gwList = make(map[uint32]*nsocket.IOClientInterface)
	server.gameList = make(map[uint64]*Game)
	server.gameIdIndex = 0

	manager := NewPlayerManager(server)
	server.playerManager = manager

	client := nsocket.NewClientSocketTcp(server)
	server.clientSocket = client
	for it := server.serverConfig.dbs.Front(); it != nil; it = it.Next() {
		cfg, ok := it.Value.(public.ServerHostConfig)
		if ok {
			client.Connect(cfg.Id, public.SERVER_TYPE_DATABASESERVER, cfg.Host)
		}
	}
	go server.clientSocket.Run()

	serverSocket := nsocket.NewServerSocketTcp(server)
	server.serverSocket = serverSocket

	// 注册帧计时器
	interval := time.Millisecond * time.Duration(server.serverConfig.frameinterval)
	server.serverSocket.RegistFrameTimer(interval, server.OnFrameCallback)

	go server.serverSocket.Run()
	server.serverSocket.Init(server.serverConfig.listenaddr)

	return nil
}

func (server *Server) OnConnectServerCallback(io *nsocket.IOServerInterface) {
	flog.GetInstance().Infof("connect success,id=%d,type=%d", (*io).GetId(), (*io).GetServerType())

	var req protocol.ReqServerRegist
	req.Serverid = server.serverConfig.serverid
	req.Servertype = public.SERVER_TYPE_GAMESERVER
	sendbuf := make([]byte, 128)
	len := protocol.ProtocolToBuffer(&req, sendbuf)
	(*io).Send(sendbuf[:len])
}

func (server *Server) OnCloseServerCallback(io *nsocket.IOServerInterface) {
	flog.GetInstance().Infof("server closed,id=%d,type=%d", (*io).GetId(), (*io).GetServerType())
}

func (server *Server) OnRecvServerCallback(io *nsocket.IOServerInterface, buf []byte) {
	var head protocol.ProtocolHead
	bio := protocol.Biostream{}
	bio.Attach(buf, len(buf))
	head.ReadHead(&bio)
	flog.GetInstance().Debugf("xyid=%d,len=%d,id=%d,type=%d", head.Xyid, head.Length, (*io).GetId(), (*io).GetServerType())

	if head.Xyid == protocol.XYID_RES_SERVERREGIST {
		var res protocol.ResServerRegist
		protocol.BufferToProtocol(buf, &res)

		if res.Flag == 0 {
			if (*io).GetServerType() == res.Servertype {
				if (*io).GetServerType() == public.SERVER_TYPE_DATABASESERVER {
					server.dbsList[(*io).GetId()] = io;
					flog.GetInstance().Infof("resServerRegist DataBaseServer success:id=%d", (*io).GetId())
				} else {
					flog.GetInstance().Warnf("unknown type,id=%d,type=%d", (*io).GetId(), (*io).GetServerType())
				}
			} else {
				flog.GetInstance().Errorf("server type error,io.id=%d,io.type=%d,res.type=%d",
					(*io).GetId(), (*io).GetServerType(), res.Servertype)
			}
		} else {
			flog.GetInstance().Errorf("resServerRegist error,flag=%d,id=%d,type=%d",
				res.Flag, (*io).GetId(), (*io).GetServerType())
		}
		return
	}

	switch (*io).GetServerType() {
	case public.SERVER_TYPE_DATABASESERVER:
		server.ReciveFromDBS(io, head, buf)
	default:
		flog.GetInstance().Warnf("unknown type,id=%d,type=%d", (*io).GetId(), (*io).GetServerType())
	}
}

func (server *Server) OnAcceptClientCallback(io *nsocket.IOClientInterface) {
	flog.GetInstance().Infof("accept a client")
}

func (server *Server) OnCloseClientCallback(io *nsocket.IOClientInterface) {
	flog.GetInstance().Infof("close a client,id=%d,type=%d", (*io).GetId(), (*io).GetClientType())
	switch (*io).GetClientType() {
	case public.SERVER_TYPE_GATEWAY:
		_, ok := server.gwList[(*io).GetId()]
		if ok {
			delete(server.gwList, (*io).GetId())
		}
	}
}

func (server *Server) OnRecvClientCallback(io *nsocket.IOClientInterface, buf []byte) {
	var head protocol.ProtocolHead
	bio := protocol.Biostream{}
	bio.Attach(buf, len(buf))
	head.ReadHead(&bio)
	flog.GetInstance().Debugf("xyid=%d,len=%d,id=%d,type=%d",
		head.Xyid, head.Length, (*io).GetId(), (*io).GetClientType())

	// 除了注册和心跳协议,其他协议不能在注册成功前接收
	if (head.Xyid != protocol.XYID_HEARTBEAT && head.Xyid != protocol.XYID_REQ_SERVERREGIST) &&
		(*io).GetClientType() == public.SERVER_TYPE_UNKNOWN {
		flog.GetInstance().Errorf("server isn't registed,can not recive xyid=%d,tempid=%d", head.Xyid, (*io).GetTempId())
		return
	}

	if head.Xyid == protocol.XYID_REQ_SERVERREGIST {
		var req protocol.ReqServerRegist
		protocol.BufferToProtocol(buf, &req)

		var res protocol.ResServerRegist
		res.Flag = 0
		res.Servertype = public.SERVER_TYPE_GAMESERVER

		if req.Servertype == public.SERVER_TYPE_GATEWAY {
			_, ok := server.gwList[req.Serverid]
			if ok {
				res.Flag = 2
			} else {
				(*io).SetId(req.Serverid)
				(*io).SetClientType(req.Servertype)
				server.gwList[(*io).GetId()] = io
			}
		} else {
			res.Flag = 3
		}

		flog.GetInstance().Infof("Recv ReqServerRegist:id=%d,type=%d,resFlag=%d", req.Serverid, req.Servertype, res.Flag)
		sendbuf := make([]byte, 128)
		len := protocol.ProtocolToBuffer(&res, sendbuf)
		(*io).Send(sendbuf[:len])
		return
	} else if head.Xyid == protocol.XYID_HEARTBEAT {
		var heart protocol.HeartBeat
		protocol.BufferToProtocol(buf, &heart)
		flog.GetInstance().Debugf("Recv heart,id=%d,type=%d,time=%d", (*io).GetId(), (*io).GetClientType(), heart.Timestamp)
		return
	}

	switch (*io).GetClientType() {
	case public.SERVER_TYPE_GATEWAY:
		server.ReciveFromGW(io, head, buf)
	default:
		flog.GetInstance().Warnf("unknown type,id=%d,type=%d", (*io).GetId(), (*io).GetClientType())
	}
}

func (server *Server) GetGameId() uint64 {
	// TODO 先简单处理
	server.gameIdIndex++
	return server.gameIdIndex
}

func (server *Server) DoMatch() {
	if server.matchList.Len() >= 2 {
		game := NewGame(server)

		item := server.matchList.Front()
		player, ok := item.Value.(*Player)
		if ok {
			game.AddPlayer(player)
		}
		server.matchList.Remove(item)

		item = server.matchList.Front()
		player, ok = item.Value.(*Player)
		if ok {
			game.AddPlayer(player)
		}
		server.matchList.Remove(item)

		ret := game.DoGameStart(server.GetGameId())
		if !ret {
			flog.GetInstance().Errorf("Game start error")
			return
		}
		server.gameList[game.Id] = game
	}
}

func (server *Server) OnTimerCallback() {
	server.DoMatch()

	for _, v := range server.dbsList {
		(*v).TryHeartBeat()
	}

	for _, v := range server.gameList {
		v.OnTimer()
	}
	for k, v := range server.gameList {
		// TODO 还不清楚怎么在循环内部清理，安全起见另起一个循环吧
		if v.IsEnd() {
			delete(server.gameList, k)
		}
	}
}

func (server *Server) OnFrameCallback() {
	for _, v := range server.gameList {
		v.OnFrameTimer()
	}
}

func (server *Server) SendToDBS(buf []byte) {
	for _, io := range server.dbsList {
		(*io).Send(buf)
		break
	}
}

func (server *Server) SendToGW(buf []byte, gwid uint32) {
	if gwid == 0 {
		for _, io := range server.gwList {
			(*io).Send(buf)
			break
		}
	} else {
		io, ok := server.gwList[gwid]
		if ok {
			(*io).Send(buf)
		} else {
			flog.GetInstance().Warnf("Can not get connection,gwid=%d", gwid)
		}
	}
}

func (server *Server) LeaveMatchingList(numid uint32) {
	for item := server.matchList.Front(); item != nil; item = item.Next() {
		player, ok := item.Value.(*Player)
		if ok {
			if player.Numid == numid {
				server.matchList.Remove(item)
				return
			}
		}
	}
}

func (server *Server) JoinMatchingList(player *Player) bool {
	server.matchList.PushBack(player)
	return true
}

func (server *Server) PlayerReConnect(player *Player) bool {
	game := player.GetGame()
	if game != nil {
		game.PlayerReconnect(player)
		return true
	}
	return false
}

func main() {
	defer public.CrashCatcher()
	server := Server{}
	err := server.init()
	if err != nil {
		fmt.Printf("Server init error:%s", err)
	}
}
