package main

import (
	"DrillServerGo/flog"
	"DrillServerGo/protocol"
	"DrillServerGo/socket"
)

func (server *Server) dbs_ResLogin(io *nsocket.IOServerInterface, head protocol.ProtocolHead, buf []byte) {
	player := server.playerManager.FindPlayerByTempid(head.Numid)
	if player != nil {
		var res protocol.ResLogin
		protocol.BufferToProtocol(buf, &res)
		flog.GetInstance().Infof("ResLogin:numid=%d,tempid=%d,flag=%d", res.Numidxy, res.Numid, res.Flag)

		if res.Flag == 0 {
			server.playerManager.LoginSuccess(player, res)
		}

		player.Send(buf)
	} else {
		flog.GetInstance().Errorf("ResLogin:Can not get player,tempid=%d", head.Numid)
	}
}

func (server *Server) dbs_ResRegist(io *nsocket.IOServerInterface, head protocol.ProtocolHead, buf []byte) {
	// 截获的原因是需要通过tempid找到相应玩家
	player := server.playerManager.FindPlayerByTempid(head.Numid)
	if player != nil {
		player.Send(buf)
	} else {
		flog.GetInstance().Errorf("ResRegist:Can not get player,tempid=%d", head.Numid)
	}
}

func (server *Server) ReciveFromDBS(io *nsocket.IOServerInterface, head protocol.ProtocolHead, buf []byte) {
	switch head.Xyid {
	case protocol.XYID_HEARTBEAT:
		flog.GetInstance().Debugf("Recv HeartBeat from DBS")
	case protocol.XYID_RES_LOGIN:
		server.dbs_ResLogin(io, head, buf)
	case protocol.XYID_RES_REGIST: // 未测
		server.dbs_ResRegist(io, head, buf)
	default:
		flog.GetInstance().Warnf("Unknown xyid,id=%d,xyid=%d", (*io).GetId(), head.Xyid)
	}
}
