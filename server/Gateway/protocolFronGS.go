package main

import (
	"DrillServerGo/flog"
	"DrillServerGo/protocol"
	"DrillServerGo/socket"
)

func (server *Server) gs_ResJoinRoom(io *nsocket.IOServerInterface, head protocol.ProtocolHead, buf []byte) {
	player := server.playerManager.FindPlayerByNumid(head.Numid)
	if player != nil {
		var res protocol.ResJoinRoom
		protocol.BufferToProtocol(buf, &res)
		flog.GetInstance().Debugf("JoinRoom end,numid=%d", head.Numid)

		if res.Flag == 0 {
			player.GSId = (*io).GetId()
		}

		player.Send(buf)
	} else {
		flog.GetInstance().Errorf("Can not get player.numid=%d", head.Numid)
	}
}

func (server *Server) ReciveFromGS(io *nsocket.IOServerInterface, head protocol.ProtocolHead, buf []byte) {
	switch head.Xyid {
	case protocol.XYID_HEARTBEAT:
		flog.GetInstance().Debugf("Recv HeartBeat from GS")
	case protocol.XYID_RES_JOINROOM:
		server.gs_ResJoinRoom(io, head, buf)
	default:
		player := server.playerManager.FindPlayerByNumid(head.Numid)
		if player != nil {
			player.Send(buf)
		} else {
			flog.GetInstance().Warnf("Can not get player,numid=%d,xyid=%d", head.Numid, head.Xyid)
		}
	}
}
