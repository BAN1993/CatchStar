package main

import (
	"DrillServerGo/flog"
	"DrillServerGo/protocol"
	"DrillServerGo/socket"
)

func (server *Server) gs_ReqJoinRoom(io *nsocket.IOClientInterface, head protocol.ProtocolHead, buf []byte) {
	var req protocol.ReqJoinRoom
	protocol.BufferToProtocol(buf, &req)
	flog.GetInstance().Infof("ReqJoinRoom:numid=%d,gwid=%d", req.Numid, req.GWid)

	// 简答处理,都认为成功
	var res protocol.ResJoinRoom
	res.Numid = req.Numid
	res.Flag = 0
	res.GWid = req.GWid
	res.Nickname = req.Nickname

	sendbuf := make([]byte, 128)
	len := protocol.ProtocolToBuffer(&res, sendbuf)
	(*io).Send(sendbuf[:len])
}

func (server *Server) ReciveFromGS(io *nsocket.IOClientInterface, head protocol.ProtocolHead, buf []byte) {
	switch head.Xyid {
	case protocol.XYID_REQ_JOINROOM:
		server.gs_ReqJoinRoom(io, head, buf)
	default:
		flog.GetInstance().Error("Unknown xyid=%d,len=%d", head.Xyid, head.Length)
		break
	}
}
