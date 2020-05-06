package main

import (
	"DrillServerGo/protocol"
	"DrillServerGo/socket"
	"DrillServerGo/flog"
)

func (server *Server) gw_ReqJoinRoom(io *nsocket.IOClientInterface, head protocol.ProtocolHead, buf []byte) {
	var req protocol.ReqJoinRoom
	protocol.BufferToProtocol(buf, &req)
	flog.GetInstance().Infof("ReqJoinRoom:numid=%d,nickname=%s,gwid=%d", req.Numid, req.Nickname, req.GWid)

	server.SendToDBS(buf)
}

func (server *Server) gw_ReqPlayerLeave(io *nsocket.IOClientInterface, head protocol.ProtocolHead, buf []byte) {
	var req protocol.ReqPlayerLeave
	protocol.BufferToProtocol(buf, &req)
	flog.GetInstance().Infof("ReqPlayerLeave:numid=%d", req.Numid)
	server.LeaveMatchingList(req.Numid)
	server.playerManager.PlayerLeave(req.Numid)
}

func (server *Server) gw_ReqOpt(io *nsocket.IOClientInterface, head protocol.ProtocolHead, buf []byte) {
	var req protocol.ReqOpt
	protocol.BufferToProtocol(buf, &req)
	flog.GetInstance().Debugf("ReqOpt:numid=%d,optcnt=%d,optlen=%d", req.Numid, req.OptCnt, req.OptLen)
	player := server.playerManager.FindPlayer(req.Numid)
	if player != nil {
		game := player.GetGame()
		if game != nil {
			game.RecordOpt(player, req)
		} else {
			flog.GetInstance().Warnf("Game is nil,numid=%d", req.Numid)
		}
	} else {
		flog.GetInstance().Warnf("Can not get player,numid=%d", req.Numid)
	}
}

func (server *Server) gw_ReqConnect(io *nsocket.IOClientInterface, head protocol.ProtocolHead, buf []byte) {
	var req protocol.ReqConnect
	protocol.BufferToProtocol(buf, &req)
	flog.GetInstance().Infof("ReqConnect:numid=%d", req.Numid)
	player := server.playerManager.FindPlayer(req.Numid)
	if player != nil {
		game := player.GetGame()
		if game != nil {
			game.PlayerConnect(player)
		} else {
			flog.GetInstance().Warnf("Game is nul,numid=%d", req.Numid)
		}
	} else {
		flog.GetInstance().Warnf("Can not get player,numid=%d", req.Numid)
	}
}

func (server *Server) ReciveFromGW(io *nsocket.IOClientInterface, head protocol.ProtocolHead, buf []byte) {
	switch head.Xyid {
	case protocol.XYID_REQ_JOINROOM:
		server.gw_ReqJoinRoom(io, head, buf)
	case protocol.XYID_REQ_PLAYER_LEAVE:
		server.gw_ReqPlayerLeave(io, head, buf)
	case protocol.XYID_REQ_OPT:
		server.gw_ReqOpt(io, head, buf)
	case protocol.XYID_REQ_CONNECT:
		server.gw_ReqConnect(io, head, buf)
	default:
		flog.GetInstance().Error("Unknown xyid=%d,len=%d", head.Xyid, head.Length)
		break
	}
}