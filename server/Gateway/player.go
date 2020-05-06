package main

import "DrillServerGo/socket"

const (
	PS_UNDEFINE = 0
	PS_WAIT_KEY = 1 // 等待协议加密密匙(TODO 先不做)
	PS_WAIT_VERSION_CHECK = 2 // 等待确认客户端版本(TODO 先不做)
	PS_VERSION_CHECK_SUCCESS = 3 // 确认版本成功
	PS_WAIT_AUTH = 4 // 等待登陆验证
	PS_AUTH_SUCCESS = 5 // 登陆成功,可以正常收发协议
)

type Player struct {
	manager *PlayerManager
	io      *nsocket.IOClientInterface

	State    uint32
	TempId   uint32
	Numid    uint32
	Nickname string
	GSId     uint32
}

func NewPlayer(manager *PlayerManager, io *nsocket.IOClientInterface, tmpid uint32) *Player {
	return &Player{
		manager: manager,
		io:      io,
		State:   PS_UNDEFINE,
		TempId:  tmpid,
		Numid:   0,
		GSId:    0,
	}
}

func (player *Player)Send(buf []byte) {
	(*player.io).Send(buf)
}
