package main

import (
	"DrillServerGo/flog"
	"DrillServerGo/protocol"
	"DrillServerGo/socket"
)

func (server *Server) dbs_ResJoinRoom(io *nsocket.IOServerInterface, head protocol.ProtocolHead, buf []byte) {
	var res protocol.ResJoinRoom
	protocol.BufferToProtocol(buf, &res)
	flog.GetInstance().Infof("ResJoinRoom:numid=%d,flag=%d,gwid=%d", res.Numid, res.Flag, res.GWid)

	player := server.playerManager.FindPlayer(res.Numid)
	if player != nil {
		// 已在房间
		// 先返回客户端进房间结果
		server.SendToGW(buf, res.GWid)

		if server.PlayerReConnect(player) {
			// 正常重连，不需要下发什么了
			return
		} else {
			// 到这里说明在房间，但还没开始游戏
			var matching protocol.NtfMatching
			matching.ProtocolHead = res.ProtocolHead
			if server.JoinMatchingList(player) {
				matching.Flag = 0
			} else {
				matching.Flag = 1
			}
			sendbuf := make([]byte, 128)
			len := protocol.ProtocolToBuffer(&matching, sendbuf)
			server.SendToGW(sendbuf[:len], res.GWid)
		}
	} else {
		// 刚进房间
		player := server.playerManager.GetNewPlayer()
		if player != nil {
			player.GWid = res.GWid
			player.Numid = res.Numid
			player.Nickname = res.Nickname
			if server.playerManager.AddPlayer(player) {
				// 先返回客户端进房间结果
				sendbuf := make([]byte, 128)
				len := protocol.ProtocolToBuffer(&res, sendbuf)
				server.SendToGW(sendbuf[:len], res.GWid)

				var matching protocol.NtfMatching
				matching.ProtocolHead = res.ProtocolHead
				if server.JoinMatchingList(player) {
					matching.Flag = 0
				} else {
					matching.Flag = 1
				}
				sendbuf = make([]byte, 128)
				len = protocol.ProtocolToBuffer(&matching, sendbuf)
				server.SendToGW(sendbuf[:len], res.GWid)
				return
			} else {
				// 添加player失败
				server.playerManager.DelPlayer(player)
				res.Flag = 2
			}
		} else {
			// 创建player失败
			res.Flag = 1
		}

		sendbuf := make([]byte, 128)
		len := protocol.ProtocolToBuffer(&res, sendbuf)
		server.SendToGW(sendbuf[:len], res.GWid)
	}
}

func (server *Server) ReciveFromDBS(io *nsocket.IOServerInterface, head protocol.ProtocolHead, buf []byte) {
	switch head.Xyid {
	case protocol.XYID_HEARTBEAT:
		flog.GetInstance().Debugf("Recv HeartBeat from DBS")
	case protocol.XYID_RES_JOINROOM:
		server.dbs_ResJoinRoom(io, head, buf)
	default:
		flog.GetInstance().Warnf("unknown xyid,id=%d,xyid=%d", (*io).GetId(), head.Xyid)
	}
}
