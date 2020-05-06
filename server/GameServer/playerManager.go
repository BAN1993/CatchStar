package main

type PlayerManager struct {
	pServer *Server
	players map[uint32]*Player // <numid,player>
}

func NewPlayerManager(server *Server) *PlayerManager {
	return &PlayerManager{
		pServer: server,
		players: make(map[uint32]*Player),
	}
}

func (manager *PlayerManager) GetNewPlayer() *Player {
	player := NewPlayer(manager.pServer)
	return player
}

func (manager *PlayerManager) AddPlayer(player *Player) bool {
	p := manager.FindPlayer(player.Numid)
	if p != nil {
		return false
	}
	manager.players[player.Numid] = player
	player.State = PLAYER_STATE_ONLINE
	return true
}

func (manager *PlayerManager) PlayerLeave(numid uint32) {
	player, ok := manager.players[numid]
	if ok {
		game := player.GetGame()
		if game == nil {
			manager.DelPlayer(player)
		} else {
			game.RecordPlayerLeave(player)
			player.State = PLAYER_STATE_OFFLINE
		}
	}
}

func (manager *PlayerManager) DelPlayer(player *Player) {
	_, ok := manager.players[player.Numid]
	if ok {
		delete(manager.players, player.Numid)
	}
}

func (manager *PlayerManager) FindPlayer(numid uint32) *Player {
	player, ok := manager.players[numid]
	if ok {
		return player
	}
	return nil
}

func (manager *PlayerManager) DoGameEnd(player *Player) {
	manager.DelPlayer(player)
}
