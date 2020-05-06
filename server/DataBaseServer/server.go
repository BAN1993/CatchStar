package main

import (
	"DrillServerGo/flog"
	"DrillServerGo/protocol"
	"DrillServerGo/public"
	"DrillServerGo/socket"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

type Server struct {
	serverConfig Configs

	gwList       map[uint32]*nsocket.IOClientInterface
	gsList       map[uint32]*nsocket.IOClientInterface
	serverSocket *nsocket.ServerSocketTcp

	mysql *sql.DB
}

func (server *Server) init() error {
	err := server.serverConfig.Load("dbs_config.ini")
	if err != nil {
		fmt.Println("Load config error:", err)
		return err
	}

	err = flog.InitLog(server.serverConfig.logfilename, server.serverConfig.loglevel)
	if err != nil {
		fmt.Println("InitLog error:", err)
		return err
	}

	server.gwList = make(map[uint32]*nsocket.IOClientInterface)
	server.gsList = make(map[uint32]*nsocket.IOClientInterface)

	// 连接数据库
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s",
		server.serverConfig.sql_user,
		server.serverConfig.sql_pwd,
		server.serverConfig.sql_host,
		server.serverConfig.sql_port,
		server.serverConfig.sql_database,
		"utf8")
	server.mysql, err = sql.Open("mysql", dsn)
	if err != nil {
		flog.GetInstance().Errorf("Open mysql error:%s,DSN:%s", err, dsn)
		return err
	}
	// 测试数据库是否连接成功
	err = server.mysql.Ping()
	if err != nil {
		flog.GetInstance().Errorf("Connect mysql error:%s", err)
		return err
	} else {
		flog.GetInstance().Infof("Connect to mysql success")
	}


	serverSocket := nsocket.NewServerSocketTcp(server)
	server.serverSocket = serverSocket

	go server.serverSocket.Run()
	server.serverSocket.Init(server.serverConfig.listenaddr)

	return nil
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
	case public.SERVER_TYPE_GAMESERVER:
		_, ok := server.gsList[(*io).GetId()]
		if ok {
			delete(server.gsList, (*io).GetId())
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
		res.Servertype = public.SERVER_TYPE_DATABASESERVER

		if req.Servertype == public.SERVER_TYPE_GATEWAY {
			_, ok := server.gwList[req.Serverid]
			if ok {
				res.Flag = 2
			} else {
				(*io).SetId(req.Serverid)
				(*io).SetClientType(req.Servertype)
				server.gwList[(*io).GetId()] = io
			}
		} else if req.Servertype == public.SERVER_TYPE_GAMESERVER {
			_, ok := server.gsList[req.Serverid]
			if ok {
				res.Flag = 2
			} else {
				(*io).SetId(req.Serverid)
				(*io).SetClientType(req.Servertype)
				server.gsList[(*io).GetId()] = io
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
	case public.SERVER_TYPE_GAMESERVER:
		server.ReciveFromGS(io, head, buf)
	default:
		flog.GetInstance().Warnf("unknown type,id=%d,type=%d", (*io).GetId(), (*io).GetClientType())
	}
}

func (server *Server) OnTimerCallback() {

}

func main() {
	defer public.CrashCatcher()
	server := Server{}
	err := server.init()
	if err != nil {
		fmt.Printf("Server init error:%s", err)
	}
}
