package main

import (
	"DrillServerGo/flog"
	"DrillServerGo/protocol"
	"DrillServerGo/public"
	"DrillServerGo/socket"
	"fmt"
)

type Server struct {
	serverConfig  Configs
	playerManager *PlayerManager

	clientSocket *nsocket.ClientSocketTcp
	serverSocket *nsocket.ServerSocketInterface
	dbsList      map[uint32]*nsocket.IOServerInterface
	gsList       map[uint32]*nsocket.IOServerInterface
}

func (server *Server) init() error {
	// 加载配置
	err := server.serverConfig.Load("gw_config.ini")
	if err != nil {
		fmt.Println("Load config error:", err)
		return err
	}

	// 初始化日志模块
	err = flog.InitLog(server.serverConfig.logfilename, server.serverConfig.loglevel)
	if err != nil {
		fmt.Println("InitLog error:", err)
		return err
	}

	// 申请内存
	server.playerManager = NewPlayerManager(server)
	server.dbsList = make(map[uint32]*nsocket.IOServerInterface)
	server.gsList = make(map[uint32]*nsocket.IOServerInterface)
	server.clientSocket = nsocket.NewClientSocketTcp(server)

	// 根据配置决定创建ws还是tcp
	var serversocket nsocket.ServerSocketInterface
	if server.serverConfig.mode == "ws" {
		flog.GetInstance().Infof("Init Web ServerSocket")
		serversocket = nsocket.NewServerSocketWeb(server)
	} else {
		flog.GetInstance().Infof("Init Tcp ServerSocket")
		serversocket = nsocket.NewServerSocketTcp(server)
	}
	server.serverSocket = &serversocket

	// 去连接DataBaseServer
	for it := server.serverConfig.dbs.Front(); it != nil; it = it.Next() {
		cfg, ok := it.Value.(public.ServerHostConfig)
		if ok {
			server.clientSocket.Connect(cfg.Id, public.SERVER_TYPE_DATABASESERVER, cfg.Host)
		}
	}

	// 去连接GameServer
	for it := server.serverConfig.gs.Front(); it != nil; it = it.Next() {
		cfg, ok := it.Value.(public.ServerHostConfig)
		if ok {
			server.clientSocket.Connect(cfg.Id, public.SERVER_TYPE_GAMESERVER, cfg.Host)
		}
	}

	go server.clientSocket.Run()
	go (*server.serverSocket).Run()

	// 初始化ServerSocket
	// 注意: 这个接口要放到最后
	(*server.serverSocket).Init(server.serverConfig.listenaddr)
	return nil
}

func (server* Server) OnConnectServerCallback(io *nsocket.IOServerInterface) {
	flog.GetInstance().Infof("connect success,id=%d,type=%d", (*io).GetId(), (*io).GetServerType())

	var req protocol.ReqServerRegist
	req.Serverid = server.serverConfig.serverid
	req.Servertype = public.SERVER_TYPE_GATEWAY
	sendbuf := make([]byte, 128)
	slen := protocol.ProtocolToBuffer(&req, sendbuf)
	(*io).Send(sendbuf[:slen])
}

func (server *Server) OnCloseServerCallback(io *nsocket.IOServerInterface) {
	flog.GetInstance().Infof("server closed,id=%d,type=%d", (*io).GetId(), (*io).GetServerType())
}

func (server *Server) OnRecvServerCallback(io *nsocket.IOServerInterface, buf []byte) {
	var head protocol.ProtocolHead
	bio := protocol.Biostream{}
	bio.Attach(buf, len(buf))
	head.ReadHead(&bio)
	flog.GetInstance().Debugf("Recv:xyid=%d,len=%d,numid=%d,id=%d,type=%d", head.Xyid, head.Length, head.Numid, (*io).GetId(), (*io).GetServerType())

	if head.Xyid == protocol.XYID_RES_SERVERREGIST {
		var res protocol.ResServerRegist
		protocol.BufferToProtocol(buf, &res)

		if res.Flag == 0 { // TODO 枚举要怎么写方便
			if (*io).GetServerType() == res.Servertype {
				if (*io).GetServerType() == public.SERVER_TYPE_DATABASESERVER {
					server.dbsList[(*io).GetId()] = io;
					flog.GetInstance().Infof("resServerRegist DataBaseServer success:id=%d", (*io).GetId())
				} else if (*io).GetServerType() == public.SERVER_TYPE_GAMESERVER {
					server.gsList[(*io).GetId()] = io;
					flog.GetInstance().Infof("resServerRegist GameServer success:id=%d", (*io).GetId())
				} else {
					flog.GetInstance().Warnf("unknown type,id=%d,type=%d", (*io).GetId(), (*io).GetServerType())
				}
			} else {
				flog.GetInstance().Errorf("server type error,io.id=%d,io.type=%d,res.type=%d", (*io).GetId(), (*io).GetServerType(), res.Servertype)
			}
		} else {
			flog.GetInstance().Errorf("resServerRegist error,flag=%d,id=%d,type=%d", res.Flag, (*io).GetId(), (*io).GetServerType())
		}
		return
	}

	switch (*io).GetServerType() {
	case public.SERVER_TYPE_DATABASESERVER:
		server.ReciveFromDBS(io, head, buf)
	case public.SERVER_TYPE_GAMESERVER:
		server.ReciveFromGS(io, head, buf)
	default:
		flog.GetInstance().Warnf("unknown type,id=%d,type=%d", (*io).GetId(), (*io).GetServerType())
	}
}

func (server *Server) OnAcceptClientCallback(io *nsocket.IOClientInterface) {
	flog.GetInstance().Infof("accept a client")
	server.playerManager.NewPlayer(io)
}

func (server *Server) OnCloseClientCallback(io *nsocket.IOClientInterface) {
	flog.GetInstance().Infof("close a client")
	server.playerManager.DelPlayer(io)
}

func (server *Server) OnRecvClientCallback(io *nsocket.IOClientInterface, buf []byte) {

	player := server.playerManager.FindPlayerByIO(io)
	if player == nil {
		flog.GetInstance().Error("Can not get player")
		(*io).Close()
		return
	}

	bio := protocol.Biostream{}
	bio.Attach(buf, len(buf))

	var head protocol.ProtocolHead
	head.ReadHead(&bio)
	flog.GetInstance().Debugf("xyid=%d,len=%d,numid=%d", head.Xyid, head.Length, head.Numid)

	if head.Xyid != protocol.XYID_HEARTBEAT &&
		head.Xyid != protocol.XYID_REQ_LOGIN {
		if player.State != PS_AUTH_SUCCESS {
			flog.GetInstance().Errorf("Player not auth,xyid=%d,state=%d", head.Xyid, player.State)
			(*io).Close()
			return
		}
	}
	head.Numid = player.Numid
	server.ReciveFromClient(player, head, buf)
}

func (server *Server) OnTimerCallback() {
	for _, v := range server.gsList {
		(*v).TryHeartBeat()
	}
	for _, v := range server.dbsList {
		(*v).TryHeartBeat()
	}
}

func (server *Server) SendToDBS(buf []byte, dbsid uint32) {
	if dbsid == 0 {
		for _, io := range server.dbsList {
			(*io).Send(buf)
			break
		}
	} else {
		io, ok := server.dbsList[dbsid]
		if ok {
			(*io).Send(buf)
		} else {
			flog.GetInstance().Warnf("Can not get connection,dbsid=%d", dbsid)
		}
	}
}

func (server *Server) SendToGs(buf []byte, gsid uint32) {
	if gsid == 0 {
		for _, io := range server.gsList {
			(*io).Send(buf)
			break
		}
	} else {
		io, ok := server.gsList[gsid]
		if ok {
			(*io).Send(buf)
		} else {
			flog.GetInstance().Warnf("Can not get connection,gsid=%d", gsid)
		}
	}
}

func main() {
	defer public.CrashCatcher()
	server := Server{}
	err := server.init()
	if err != nil {
		fmt.Printf("Server init error:%s", err)
	}
}
