package main

import (
	"DrillServerGo/flog"
	"DrillServerGo/protocol"
)

func (server *Server) client_Heartbeat(player *Player, head protocol.ProtocolHead, buf []byte) {
	var req protocol.HeartBeat
	protocol.BufferToProtocol(buf, &req)
	flog.GetInstance().Debugf("recv:timestamp=%d,tempid=%d,numid=%d", req.Timestamp, player.TempId, player.Numid)

	// 不回包
	//var resp protocol.HeartBeat
	//resp.Timestamp = uint32(time.Now().Unix())
	//flog.GetInstance().Debugf("send:timestamp=%d", resp.Timestamp)
	//sendbuf := make([]byte, 128)
	//len := protocol.ProtocolToBuffer(&resp, sendbuf)
	//player.Send(sendbuf[:len])
}

func (server *Server) client_ReqLogin(player *Player, head protocol.ProtocolHead, buf []byte) {
	var req protocol.ReqLogin
	protocol.BufferToProtocol(buf, &req)
	flog.GetInstance().Debugf("Account=%s,password=%s", req.Account, req.Password)

	if player.State == PS_VERSION_CHECK_SUCCESS ||
			player.State == PS_WAIT_AUTH {
		player.State = PS_WAIT_AUTH
		req.Numid = player.TempId // 用零时id代替，用于返回res时能找到player
		sendbuf := make([]byte, 128)
		len := protocol.ProtocolToBuffer(&req, sendbuf)
		server.SendToDBS(sendbuf[:len], 0)
	} else {
		flog.GetInstance().Errorf("State error,state=%d", player.State)

		var res protocol.ResLogin
		res.Flag = 5
		sendbuf := make([]byte, 128)
		len := protocol.ProtocolToBuffer(&res, sendbuf)
		player.Send(sendbuf[:len])
	}
}

func (server *Server) client_ReqJoinRoom(player *Player, head protocol.ProtocolHead, buf []byte) {
	var req protocol.ReqJoinRoom
	protocol.BufferToProtocol(buf, &req)
	req.Nickname = player.Nickname
	req.GWid = server.serverConfig.serverid
	flog.GetInstance().Debugf("numid=%d,nickname=%s,len=%d", req.Numid, req.Nickname, req.Length)

	sendbuf := make([]byte, 128)
	len := protocol.ProtocolToBuffer(&req, sendbuf)
	server.SendToGs(sendbuf[:len], 0) // TODO 这里先随机挑一个gs,没有做分配逻辑
}

func (server *Server) client_ReqRegist(player *Player, head protocol.ProtocolHead, buf []byte) {
	// 截获是因为包头的numid需要置为tempid
	var req protocol.ReqRegist
	protocol.BufferToProtocol(buf, &req)

	req.Numid = player.TempId

	sendbuf := make([]byte, 128)
	len := protocol.ProtocolToBuffer(&req, sendbuf)
	server.SendToDBS(sendbuf[:len], 0)
}

func (server *Server) ReciveFromClient(player *Player, head protocol.ProtocolHead, buf []byte) {
	switch head.Xyid {
	case protocol.XYID_HEARTBEAT:
		server.client_Heartbeat(player, head, buf)
	case protocol.XYID_REQ_LOGIN:
		server.client_ReqLogin(player, head, buf)
	case protocol.XYID_REQ_JOINROOM:
		server.client_ReqJoinRoom(player, head, buf)
	case protocol.XYID_REQ_OPT: // 由于go的switch和c不一样,case自带break,有没有更好的办法
		server.SendToGs(buf, player.GSId)
	case protocol.XYID_REQ_CONNECT:
		server.SendToGs(buf, player.GSId)
	case protocol.XYID_REQ_REGIST: // 未测
		server.client_ReqRegist(player, head, buf)
	case protocol.XYID_CHECK_RTT:
		player.Send(buf) // 直接回包
	default:
		flog.GetInstance().Warnf("Unknown xyid=%d,numid=%d", head.Xyid, player.Numid)
		break
	}
}