package main

import (
	"DrillServerGo/protocol"
	"DrillServerGo/socket"
)

type PlayerManager struct {
	pServer     *Server
	players     map[uint32]*Player
	numids      map[uint32]uint32
	tempIdIndex uint32
}

func NewPlayerManager(server *Server) *PlayerManager {
	return &PlayerManager{
		pServer:     server,
		players:     make(map[uint32]*Player),
		numids:      make(map[uint32]uint32),
		tempIdIndex: 0,
	}
}

func (m *PlayerManager) GetNewTempId() uint32 {
	m.tempIdIndex++
	return m.tempIdIndex
}

func (m *PlayerManager) NewPlayer(io *nsocket.IOClientInterface) {
	player := NewPlayer(m, io, m.GetNewTempId())
	(*io).SetTempId(player.TempId)
	m.players[player.TempId] = player

	// 先略去验证加密这一步，直接设置状态为等待登录吧
	player.State = PS_VERSION_CHECK_SUCCESS
}

func (m *PlayerManager) DelPlayer(io *nsocket.IOClientInterface) {
	player := m.FindPlayerByIO(io)
	if player != nil {
		if player.GSId != 0 {
			var req protocol.ReqPlayerLeave
			req.Numid = player.Numid
			sendbuf := make([]byte, 128)
			len := protocol.ProtocolToBuffer(&req, sendbuf)
			m.pServer.SendToGs(sendbuf[:len], player.GSId)
		}
		delete(m.numids, player.Numid)
		delete(m.players, player.TempId)
	}
}

func (m *PlayerManager) LoginSuccess(player *Player, res protocol.ResLogin) {
	m.numids[res.Numidxy] = player.TempId
	player.State = PS_AUTH_SUCCESS
	player.Nickname = res.Nickname
	player.Numid = res.Numidxy
}

func (m *PlayerManager) FindPlayerByIO(io *nsocket.IOClientInterface) *Player {
	return m.FindPlayerByTempid((*io).GetTempId())
}

func (m *PlayerManager) FindPlayerByTempid(tmpid uint32) *Player {
	c, ok := m.players[tmpid]
	if ok {
		return c
	}
	return nil
}

func (m *PlayerManager) FindPlayerByNumid(numid uint32) *Player {
	tempid, ok := m.numids[numid]
	if !ok {
		return nil
	}
	return m.FindPlayerByTempid(tempid)
}
