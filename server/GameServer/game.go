package main

import (
	"DrillServerGo/flog"
	"DrillServerGo/protocol"
	"container/list"
	"math/rand"
	"time"
	"unsafe"
)

const (
	GAME_STATE_WAIT_CONNECT = 0
	GAME_STATE_GAMEING = 1
	GAME_STATE_END = 2
)

type OptRecord struct {
	Numid uint32
	Opttype uint32
	Datalen uint32
	Data []byte
}

type OptList struct {
	Datas list.List
}

type Game struct {
	pServer    *Server
	Id         uint64
	StartCnt   uint32
	GameState  uint32
	StartTime  uint32
	ConnectCnt uint32

	players    map[uint32]*Player
	historyOpt map[uint32]*OptList
	tempOpt    map[uint32]*OptList

	nowFrameIndex uint32
	historyLen uint32
	historyCnt uint32

	lastCreatStarTime uint32
	starIndex uint32
	starMap map[uint32]int32
}

func NewGame(server *Server) *Game {
	return &Game{
		pServer:    server,
		Id:         0,
		StartCnt:   2,
		GameState:  GAME_STATE_WAIT_CONNECT,
		StartTime:  0,
		ConnectCnt: 0,

		players:    make(map[uint32]*Player),
		historyOpt: make(map[uint32]*OptList),
		tempOpt:    make(map[uint32]*OptList),
		nowFrameIndex:0,
		historyLen: 0,
		historyCnt: 0,

		lastCreatStarTime:0,
		starIndex:0,
		starMap:make(map[uint32]int32),
	}
}

func (game *Game) OnTimer() {
	if game.GameState != GAME_STATE_GAMEING {
		return
	}

	flog.GetInstance().Infof("id=%d,historylen=%d,historycnt=%d",
		game.Id, game.historyLen, game.historyCnt)

	// 这个要做成配置
	var maxGameTime uint32 = 180

	var now uint32 = uint32(time.Now().Unix())
	if (now-game.StartTime) > maxGameTime || game.isAllOffline() {
		flog.GetInstance().Debugf("start=%d,now=%d,need end", game.StartTime, now)
		game.doGameEnd()
	}
	if now - game.lastCreatStarTime >= 5 || game.lastCreatStarTime == 0 {
		// 创建星星
		game.createStar()
		game.lastCreatStarTime = now
	}

	// 清理超时的星星
	game.delTimeOutStar()
}

func (game *Game) OnFrameTimer() {
	if game.GameState != GAME_STATE_GAMEING {
		return // 按道理不会有这种情况
	}
	if len(game.tempOpt) <= 0 {
		return
	}

	bcnt := game.sendMapOptData(&game.tempOpt, nil)
	flog.GetInstance().Debugf("Broadcast Opt cnt=%d", bcnt)

	// 发完后把data塞入history列表, 并清理temp
	for index, opts := range game.tempOpt {
		if uint32(opts.Datas.Len()) <= 0 {
			continue
		}
		for it := opts.Datas.Front(); it != nil; it = it.Next() {
			opt, ok := it.Value.(OptRecord)
			if ok {
				game.pushHisoryOptRecord(index, opt)
			} else {
				flog.GetInstance().Errorf("is not OptRecode,index=%d", index)
			}
		}
	}
	game.tempOpt = make(map[uint32]*OptList)
}

func (game *Game) DoGameStart(id uint64) bool {
	if len(game.players) != int(game.StartCnt) {
		return false
	}
	flog.GetInstance().Debugf("Game start,id=%d", id)

	game.Id = id
	game.GameState = GAME_STATE_WAIT_CONNECT

	for _, player := range game.players {
		player.ResetGameData()
		player.SetGame(game)
	}

	game.broadcastGameStart()
	return true
}

func (game *Game) doGameEnd() {
	flog.GetInstance().Debugf("Game end.id=%d", game.Id)

	game.GameState = GAME_STATE_END
	game.broadcastGameEnd()

	for _, player := range game.players {
		game.pServer.playerManager.DoGameEnd(player)
	}

	game.clearAllData()
}

func (game *Game) IsEnd() bool {
	return game.GameState == GAME_STATE_END
}

func (game *Game) AddPlayer(player *Player) bool {
	_, ok := game.players[player.Numid]
	if ok {
		return false
	}
	game.players[player.Numid] = player
	return true
}

func (game *Game) PlayerConnect(player *Player) {
	// 先简单处理,不考虑有人没connect或者有人多次connect
	game.ConnectCnt++
	if game.ConnectCnt == uint32(len(game.players)) {
		game.GameState = GAME_STATE_GAMEING
		game.StartTime = uint32(time.Now().Unix())
		flog.GetInstance().Infof("Game start,time=%d", game.StartTime)
		game.recordGameStart()
	}
}

func (game *Game) PlayerReconnect(player *Player) {
	var ntf protocol.NtfGameStart
	ntf.Numid = player.Numid // TODO 发协议要想办法优化下,看能不能用一个接口处理包头
	sendbuf := make([]byte, 128)
	len := protocol.ProtocolToBuffer(&ntf, sendbuf)
	player.SendToGW(sendbuf[:len])

	game.recordPlayerReconnect(player)
	game.sendHistory(player)
	player.State = PLAYER_STATE_ONLINE
}

func (game *Game) clearAllData() {
	game.tempOpt = make(map[uint32]*OptList)
	game.historyOpt = make(map[uint32]*OptList)
	game.players = make(map[uint32]*Player)
	game.starMap = make(map[uint32]int32)
}

func (game *Game) isAllOffline() bool {
	for _, player := range game.players {
		if player.State == PLAYER_STATE_ONLINE {
			return false
		}
	}
	return true
}

func (game *Game) recordGameStart() {
	var record OptRecord
	record.Numid = 0
	record.Opttype = OPT_TYPE_GAME_START
	record.Data = make([]byte, 256)

	// 身份值,先简单处理
	var whoIndex uint16 = 0

	var opt OptGameStart
	opt.PlayerCnt = uint32(len(game.players))
	for _, player := range game.players {
		var p OptGameStart_PlayerData
		p.Numid = player.Numid
		p.Who = whoIndex
		whoIndex++
		p.Nickname = player.Nickname
		p.Gold = player.GameGold
		p.py = -100
		if(p.Who==0){
			p.px = -250
		}else{
			p.px = 250
		}
		flog.GetInstance().Debugf("numid=%d,who=%d,nickname=%s,gold=%d,px=%d,py=%d",
			p.Numid,p.Who,p.Nickname,p.Gold,p.px,p.py)
		opt.players.PushBack(p)
	}

	record.Datalen = protocol.ProtocolToBuffer(&opt, record.Data)
	game.pushTempOptRecord(game.nowFrameIndex, record)
}

func (game *Game) RecordPlayerLeave(player *Player) {
	var record OptRecord
	record.Numid = player.Numid
	record.Opttype = OPT_TYPE_PLAYER_LEAVE
	record.Data = make([]byte, 128)

	var opt OptEmpty
	record.Datalen = protocol.ProtocolToBuffer(&opt, record.Data)
	game.pushTempOptRecord(game.nowFrameIndex, record)
}

func (game *Game) RecordOpt(player *Player, req protocol.ReqOpt) {
	var bio protocol.Biostream
	optdata := []byte(req.Optdatas)
	bio.Attach(optdata, int(req.OptLen))

	for i := 0; i < int(req.OptCnt); i++ {
		var record OptRecord
		record.Numid = bio.ReadUint32()
		record.Opttype = bio.ReadUint32()
		record.Datalen = bio.ReadUint32()
		record.Data = bio.ReadBytes(record.Datalen)

		// 记录玩家帧号
		if player.FrameIndex<req.FrameIndex {
			player.FrameIndex = req.FrameIndex
		} else if player.FrameIndex > req.FrameIndex {
			// 这里要不要抛弃协议或者做什么处理?
			flog.GetInstance().Warnf("FrameIndex may error,numid=%d,now.index=%d,msg.index=%d", player.Numid, player.FrameIndex, req.FrameIndex)
		}

		// 游戏帧号用最大的玩家帧号
		if game.nowFrameIndex<player.FrameIndex {
			game.nowFrameIndex = player.FrameIndex
		}

		game.pushTempOptRecord(req.FrameIndex, record)
	}
}

func (game *Game) recordPlayerReconnect(player *Player) {
	var record OptRecord
	record.Numid = player.Numid
	record.Opttype = OPT_TYPE_PLAYER_RECONNECT
	record.Data = make([]byte, 128)

	var opt OptEmpty
	record.Datalen = protocol.ProtocolToBuffer(&opt, record.Data)
	game.pushTempOptRecord(game.nowFrameIndex, record)
}

func (game *Game) pushHisoryOptRecord(index uint32, opt OptRecord) {
	optlist, ok := game.historyOpt[index]
	if ok {
		optlist.Datas.PushBack(opt)
	} else {
		var l OptList
		l.Datas.PushBack(opt)
		game.historyOpt[index] = &l
	}
	game.historyLen += uint32(unsafe.Sizeof(opt)) // TODO 这个是否准确有待考证
	game.historyCnt++
}

func (game *Game) pushTempOptRecord(index uint32, opt OptRecord) {
	if !game.checkOpt(opt) {
		flog.GetInstance().Errorf("checkOpt error,index=%d,opttype=%d", index, opt.Opttype)
		return
	}
	optlist, ok := game.tempOpt[index]
	if ok {
		optlist.Datas.PushBack(opt)
	} else {
		var l OptList
		l.Datas.PushBack(opt)
		game.tempOpt[index] = &l
	}
}

func (game *Game) broadcastGameStart() {
	for _, player := range game.players {
		if player.State == PLAYER_STATE_ONLINE {
			var ntf protocol.NtfGameStart
			ntf.Numid = player.Numid
			sendbuf := make([]byte, 128)
			len := protocol.ProtocolToBuffer(&ntf, sendbuf)
			flog.GetInstance().Debugf("Borad GameStart,numid=%d,gwid=%d", ntf.Numid, player.GWid)

			player.SendToGW(sendbuf[:len])
		}
	}
}

func (game *Game) broadcastGameEnd() {
	var ntf protocol.NtfGameEnd
	sendbuf := make([]byte, 128)

	for _, player := range game.players {
		if player.State == PLAYER_STATE_ONLINE {
			ntf.Numid = player.Numid
			len := protocol.ProtocolToBuffer(&ntf, sendbuf)
			player.SendToGW(sendbuf[:len])
		}
	}
}

func (game *Game) broadcastOpt(ntf protocol.NtfOpt, player *Player) {

	if player == nil {
		var sendcnt uint32 = 0
		for _, p := range game.players {
			if p.State == PLAYER_STATE_ONLINE {
				ntf.Numid = p.Numid
				sendbuf := make([]byte, 5120) // TODO 这里的长度会有问题,后面要优化发送协议的接口
				len := protocol.ProtocolToBuffer(&ntf, sendbuf)

				p.SendToGW(sendbuf[:len])
				sendcnt++
			}
		}
		flog.GetInstance().Debugf("broadcastOpt:flag=%d,framecnt=%d,sendcnt=%d", ntf.Flag, ntf.FrameCnt, sendcnt)
	} else {
		ntf.Numid = player.Numid
		sendbuf := make([]byte, 5120) // TODO 这里的长度会有问题,后面要优化发送协议的接口
		len := protocol.ProtocolToBuffer(&ntf, sendbuf)

		player.SendToGW(sendbuf[:len])
		flog.GetInstance().Debugf("broadcastOpt:flag=%d,framecnt=%d,tonumid=%d", ntf.Flag, ntf.FrameCnt, player.Numid)
	}
}

func (game *Game) sendHistory(player *Player) {
	bcnt := game.sendMapOptData(&game.historyOpt, player)
	flog.GetInstance().Infof("Send to player history cnt=%d", bcnt)
}

func (game *Game) sendMapOptData(datas *map[uint32]*OptList, player *Player) uint32 {
	var ntf protocol.NtfOpt
	ntf.Framedatas = make([]byte, xc_max_framedatas_len)
	bio := protocol.Biostream{}
	bio.Attach(ntf.Framedatas, len(ntf.Framedatas))

	var retcnt uint32 = 0
	var maxlen uint32 = xc_max_framedatas_len - protocol.HeadLen - 100 // 安全起见,多预留一些位置
	var nowframecnt uint32 = 0
	var nowoptcnt uint32 = 0
	var nowframebeginoffset uint32 = 0

	// 结构体里没有定义,但是实际结构如下(framedata):
	//	flag
	//	framecnt
	//	framelen
	//	framedatas{ --这里下面开始就非协议定义的字段了
	//		frameindex
	//		optcnt
	//		optdatas{
	//			numid
	//			opttype
	//			datalen
	//			data
	//		}
	//	}

	for index, optlist := range *datas {
		if optlist.Datas.Len() <= 0 {
			continue
		}
		nowframebeginoffset = bio.GetOffset()
		bio.WriteUint32(index)
		bio.WriteUint32(0) // 这里流的是optcnt,这时候还不知道,后面再填充
		nowframecnt++
		nowoptcnt = 0

		for it := optlist.Datas.Front(); it != nil; it = it.Next() {
			record, ok := it.Value.(OptRecord)
			if ok {
				// unsafe.Sizeof 是否能获取真正的大小?
				if (bio.GetOffset() + uint32(unsafe.Sizeof(record))) >= maxlen {
					ntf.Flag = 0 // continue
					ntf.FrameCnt = nowframecnt
					ntf.FrameLen = bio.GetOffset()

					// 填充opt数目
					bio.Seek(nowframebeginoffset+4) // 距离头部一个int
					bio.WriteUint32(nowoptcnt)
					bio.Seek(ntf.FrameLen)

					game.broadcastOpt(ntf, player)

					ntf = protocol.NtfOpt{}
					ntf.Framedatas = make([]byte, xc_max_framedatas_len)
					bio = protocol.Biostream{}
					bio.Attach(ntf.Framedatas, len(ntf.Framedatas))

					nowoptcnt = 0
					nowframecnt = 0
					nowframebeginoffset = 0

					bio.WriteUint32(index)
					bio.WriteUint32(0) // 这里流的是optcnt,这时候还不知道,后面再填充
					nowframecnt++
				}
				nowoptcnt++

				bio.WriteUint32(record.Numid)
				bio.WriteUint32(record.Opttype)
				bio.WriteUint32(record.Datalen)
				bio.WriteBytes(record.Data, record.Datalen)
				retcnt++
			} else {
				panic("RecordMap error,is not OptRecord")
			}
		}

		// 填充opt数目
		ntf.FrameLen = bio.GetOffset()
		bio.Seek(nowframebeginoffset+4) // 距离头部一个int
		bio.WriteUint32(nowoptcnt)
		bio.Seek(ntf.FrameLen)
	}
	if bio.GetOffset() > 0 {
		ntf.Flag = 1 // end
		ntf.FrameCnt = nowframecnt
		ntf.FrameLen = bio.GetOffset()

		// 填充帧数目
		bio.Seek(nowframebeginoffset+4) // 距离头部一个int
		bio.WriteUint32(nowoptcnt)
		bio.Seek(ntf.FrameLen)

		game.broadcastOpt(ntf, player)
	}

	return retcnt
}

func (game *Game) createStar() {
	// 告知客户端生成了一个星星，位置由服务端控制
	var record OptRecord
	record.Numid = 0
	record.Opttype = OPT_TYPE_NEW_STAR
	record.Data = make([]byte, 128)

	var opt OptNewStar
	game.starIndex++
	opt.index = game.starIndex
	// TODO 位置先写死吧
	opt.px = rand.Int31n(960) - 960/2
	opt.py = -128 + 50 + rand.Int31n(150)
	opt.duration = rand.Uint32() % 5 + 1
	flog.GetInstance().Infof("Create star:i=%d,x=%d,y=%d,duration=%d", opt.index, opt.px, opt.py, opt.duration)
	record.Datalen = protocol.ProtocolToBuffer(&opt, record.Data)
	game.pushTempOptRecord(game.nowFrameIndex, record) // TODO 现在都是0
}

func (game *Game) delStar(index uint32) {
	var record OptRecord
	record.Numid = 0
	record.Opttype = OPT_TYPE_DEL_STAR
	record.Data = make([]byte, 128)

	var opt OptDelStar
	opt.index = index
	flog.GetInstance().Infof("Del star:i=%d", index)
	record.Datalen = protocol.ProtocolToBuffer(&opt, record.Data)
	game.pushTempOptRecord(game.nowFrameIndex, record) // TODO 现在都是0
}

func (game *Game) delTimeOutStar() {
	for index, duration := range game.starMap {
		if duration == 0 { // 防止自己多次删除？
			game.delStar(index)
		} else {
			game.starMap[index]--
		}
	}
}

/**
 * 截取opt，如果要对请求做校验或逻辑要放在这里
 * 返回true才会压入opt队列
 */
func (game *Game) checkOpt(record OptRecord) bool {
	switch record.Opttype {
	case OPT_TYPE_NEW_STAR:
		return game.checkOpt_NewStar(record)
	case OPT_TYPE_CATCH_STAR:
		return game.checkOpt_CatchStar(record)
	case OPT_TYPE_DEL_STAR:
		return game.checkOpt_DelStar(record)
	}
	return true
}

func (game *Game) checkOpt_NewStar(record OptRecord) bool {
	var opt OptNewStar
	protocol.BufferToProtocol(record.Data, &opt)

	_, had := game.starMap[opt.index]
	if had {
		flog.GetInstance().Errorf("Had star index=%d", opt.index)
		return false
	}
	game.starMap[opt.index] = int32(opt.duration)
	return true
}

func (game *Game) checkOpt_CatchStar(record OptRecord) bool {
	var opt OptCatchStar
	protocol.BufferToProtocol(record.Data, &opt)

	_, had := game.starMap[opt.index]
	if !had {
		flog.GetInstance().Errorf("Had not star index=%d", opt.index)
		return false
	}
	delete(game.starMap, opt.index)
	return true
}

func (game *Game) checkOpt_DelStar(record OptRecord) bool {
	var opt OptDelStar
	protocol.BufferToProtocol(record.Data, &opt)

	_, had := game.starMap[opt.index]
	if !had {
		flog.GetInstance().Errorf("Had not star index=%d", opt.index)
		return false
	}
	delete(game.starMap, opt.index)
	return true
}
