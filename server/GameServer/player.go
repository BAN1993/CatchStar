package main

const (
	PLAYER_STATE_UNKNOWN = 0
	PLAYER_STATE_ONLINE = 1
	PLAYER_STATE_OFFLINE = 2
)

type Player struct {
	pServer     *Server
	Numid       uint32
	Nickname    string
	GWid        uint32
	State       uint32
	pGame       *Game
	OfflineTime uint32

	// 游戏内数据
	GameGold   uint32
	FrameIndex uint32
}

func NewPlayer(server *Server) *Player {
	return &Player{
		pServer:     server,
		Numid:       0,
		Nickname:    "",
		GWid:        0,
		State:       PLAYER_STATE_UNKNOWN,
		pGame:       nil,
		OfflineTime: 0,
		GameGold:    0,
		FrameIndex:  0,
	}
}

func (player *Player) ResetGameData() {
	player.GameGold = 2000
	player.FrameIndex = 0
}

func (player *Player) SetGame(game *Game) {
	player.pGame = game
}

func (player *Player) GetGame() *Game {
	return player.pGame
}

func (player *Player) SendToGW(buf []byte) {
	player.pServer.SendToGW(buf, player.GWid)
}
