package main

import (
	"DrillServerGo/protocol"
	"container/list"
)

const (
	OPT_TYPE_GAME_START = 0
	OPT_TYPE_PLAYER_LEAVE = 1
	OPT_TYPE_PLAYER_RECONNECT = 2
	OPT_TYPE_MOVE = 3

	OPT_TYPE_CLIENT_BEGIN = 3000
	OPT_TYPE_CREATE_MONSTER = 3001
	OPT_TYPE_CREATE_TURRET = 3002
	OPT_TYPE_TURRET_SHOOT = 3003
	OPT_TYPE_NEW_STAR = 3006
	OPT_TYPE_CATCH_STAR = 3007
	OPT_TYPE_CLIENT_POSITION = 3008
	OPT_TYPE_DEL_STAR = 3009
)

const (
	xc_max_player_cnt = 20
	xc_max_framedatas_len = 4096
)

type OptEmpty struct {
	protocol.ProtocolInterface
}
func (this *OptEmpty) Read(bio *protocol.Biostream) {

}
func (this *OptEmpty) Write(bio *protocol.Biostream) {

}

type OptGameStart struct {
	protocol.ProtocolInterface

	PlayerCnt uint32
	players   list.List // OptGameStart_PlayerData
}
type OptGameStart_PlayerData struct {
	Who      uint16
	Numid    uint32
	Nickname string
	Gold     uint32
	px       int32
	py       int32
}
func (this *OptGameStart) Read(bio *protocol.Biostream) {
	this.PlayerCnt = bio.ReadUint32()
	for i := 0; i < int(this.PlayerCnt); i++ {
		var p OptGameStart_PlayerData
		p.Who = bio.ReadUint16()
		p.Numid = bio.ReadUint32()
		p.Nickname = bio.ReadString()
		p.Gold = bio.ReadUint32()
		p.px = bio.ReadInt32()
		p.py = bio.ReadInt32()
		this.players.PushBack(p)
	}
}
func (this *OptGameStart) Write(bio *protocol.Biostream) {
	bio.WriteUint32(uint32(this.players.Len()))
	for it := this.players.Front(); it != nil; it = it.Next() {
		p, ok := it.Value.(OptGameStart_PlayerData)
		if ok {
			bio.WriteUint16(p.Who)
			bio.WriteUint32(p.Numid)
			bio.WriteString(p.Nickname)
			bio.WriteUint32(p.Gold)
			bio.WriteInt32(p.px)
			bio.WriteInt32(p.py)
		} else {
			panic("OptGameStart write error")
		}
	}
}

type OptNewStar struct {
	protocol.ProtocolInterface
	index uint32
	px int32
	py int32
	duration uint32
}
func (this *OptNewStar) Read(bio *protocol.Biostream) {
	this.index = bio.ReadUint32()
	this.px = bio.ReadInt32()
	this.py = bio.ReadInt32()
	this.duration = bio.ReadUint32()
}
func (this *OptNewStar) Write(bio *protocol.Biostream) {
	bio.WriteUint32(this.index)
	bio.WriteInt32(this.px)
	bio.WriteInt32(this.py)
	bio.WriteUint32(this.duration)
}

type OptCatchStar struct {
	protocol.ProtocolInterface
	index uint32
}
func (this *OptCatchStar) Read(bio *protocol.Biostream) {
	this.index = bio.ReadUint32()
}

type OptDelStar struct {
	protocol.ProtocolInterface
	index uint32
}
func (this *OptDelStar) Read(bio *protocol.Biostream) {
	this.index = bio.ReadUint32()
}
func (this *OptDelStar) Write(bio *protocol.Biostream) {
	bio.WriteUint32(this.index)
}
