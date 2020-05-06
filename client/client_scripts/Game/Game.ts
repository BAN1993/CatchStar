import { NetWork, NetWorkState } from "../Net/NetWork"
import { msgReqConnect, msgNtfOpt, msgReqOpt, msgCheckRtt } from "../Protocol/ProtocolMsg"
import { ProtocolHead } from "../Protocol/ProtocolHead"
import { XYID } from "../Protocol/ProtocolXYID"
import { biostream } from "../Protocol/biostream"
import { FrameData, OPTTYPE, OptData, OptGameStart, OptNewStar, OptCatchStar, OptDelStar, OptMove, OptJump } from "./Opt"
import { EnemyData } from "../Data/EnemyData"
import { MyData } from "../Data/MyData"
import { KeyDefine, GameConfig, MoveType } from "../SomeDefine"
import { DataPosters } from "../Data/DataPoster"

const { property, ccclass } = cc._decorator

@ccclass
export class Game extends cc.Component {
  @property(cc.Prefab)
  private starPrefab: cc.Prefab = null
  @property(cc.Prefab)
  private playerPrefab: cc.Prefab = null
  @property(cc.Prefab)
  private scorePrefab: cc.Prefab = null
  @property(cc.Label)
  private rtttips: cc.Label = null

  private gamestart: boolean = false
  private leftOrRight: number = MoveType.Stop
  private keyDownCount: number = 0
  private netFrameId: number = 0
  private updateFrameIndex: number = 0

  protected onLoad() {
    cc.view.setOrientation(cc.macro.ORIENTATION_LANDSCAPE)
    // 切场景成功后告诉服务端
    NetWork.getInstance().regist(this)
    cc.log("[Game]Send reqConnect")
    let req = new msgReqConnect()
    NetWork.getInstance().sendMsg(req)

    // 初始化输入监听
    this.addEventListeners()

    this.netFrameId = 0
    this.updateFrameIndex = 0
    this.registData()

    this.rtttips.string = "RTT:?"
  }

  protected update(dt: number) {
    this.updateFrameIndex++
    if (this.updateFrameIndex % 2 == 0) { // 两帧更新一次位置
      this.netFrameId++
    }
    if (this.updateFrameIndex % 15 == 0) {
      this.checkRTT()
    }
  }
  /************************************************************************************************************
   ************************************************************************************************************
   ************************************************************************************************************
   ************************************************************************************************************
   ************************************************************************************************************
   自定义接口
   */

  public getNowFrameId(): number {
    return this.netFrameId
  }

  private selfMoveLeft() {
    if (this.leftOrRight != MoveType.Left) {
      this.leftOrRight = MoveType.Left
      this.keyDownCount++
      //this.sendOptMove(this.leftOrRight)
      MyData.getInstance().data.updateAim(this.leftOrRight)
    }
  }

  private selMoveRight() {
    if (this.leftOrRight != MoveType.Right) {
      this.leftOrRight = MoveType.Right
      this.keyDownCount++
      //this.sendOptMove(this.leftOrRight)
      MyData.getInstance().data.updateAim(this.leftOrRight)
    }
  }

  private selfStopMove() {
    this.keyDownCount--
    if (this.keyDownCount <= 0) {
      this.leftOrRight = MoveType.Stop
      //this.sendOptMove(this.leftOrRight)
      MyData.getInstance().data.updateAim(this.leftOrRight)
    }
  }

  public selfJump() {
    let menode = this.node.getChildByName(KeyDefine.Key_Node + KeyDefine.Key_Me)
    if (menode.getComponent("Player").getCanJump()) {
      this.sendOptJump()
      MyData.getInstance().data.updateJump(menode.getPosition())
    }
  }

  /************************************************************************************************************
   ************************************************************************************************************
   ************************************************************************************************************
   ************************************************************************************************************
   ************************************************************************************************************
   玩家操作相关接口
   */

  private addEventListeners() {
    cc.systemEvent.on(cc.SystemEvent.EventType.KEY_DOWN, this.onKeyDown, this)
    cc.systemEvent.on(cc.SystemEvent.EventType.KEY_UP, this.onKeyUp, this)
    cc.find("Canvas").on(cc.Node.EventType.TOUCH_START, this.onScreenTouchStart, this)
    cc.find("Canvas").on(cc.Node.EventType.TOUCH_CANCEL, this.onScreenTouchEnd, this)
    cc.find("Canvas").on(cc.Node.EventType.TOUCH_END, this.onScreenTouchEnd, this)
  }

  private onKeyDown(event: cc.Event.EventKeyboard) {
    switch ((event as any).keyCode) {
      case cc.macro.KEY.a:
      case cc.macro.KEY.left:
        this.selfMoveLeft()
        break
      case cc.macro.KEY.d:
      case cc.macro.KEY.right:
        this.selMoveRight()
        break
      case cc.macro.KEY.space:
        this.selfJump()
        break
    }
  }

  private onKeyUp(event: cc.Event.EventKeyboard) {
    switch ((event as any).keyCode) {
      case cc.macro.KEY.a:
      case cc.macro.KEY.left:
        this.selfStopMove()
        break
      case cc.macro.KEY.d:
      case cc.macro.KEY.right:
        this.selfStopMove()
        break
    }
  }

  private onScreenTouchStart(event: cc.Event.EventTouch) {
    if (event.getLocationX() > cc.winSize.width / 2) {
      this.selMoveRight()
    } else {
      this.selfMoveLeft()
    }
  }

  private onScreenTouchEnd() {
    this.selfStopMove()
  }

  /************************************************************************************************************
   ************************************************************************************************************
   ************************************************************************************************************
   ************************************************************************************************************
   ************************************************************************************************************
   数据更新
   */

  private registData() {
    DataPosters.getInstance().regist(KeyDefine.Key_CTL + KeyDefine.Key_Me, this)
  }

  public onDataUpdate(key: string, message: any) {
    cc.log("[Game][onDataUpdate]key=%s,message=", key, message)
    if (key == KeyDefine.Key_CTL + KeyDefine.Key_Me) {
      this.selfJump()
    }
  }


  /************************************************************************************************************
   ************************************************************************************************************
   ************************************************************************************************************
   ************************************************************************************************************
   ************************************************************************************************************
   发送网络消息相关接口
   */

  // public sendOptPosition(aim: number, pos: cc.Vec2) {
  //   let opt = new OptPosition()
  //   opt.aim = aim
  //   opt.px = pos.x
  //   opt.py = pos.y
  //   cc.log("[sendOptPosition]aim=%d,x=%d,y=%d", opt.aim, opt.px, opt.py)
  //   this.sendOptData(OPTTYPE.CLIENT_POSITION, opt, 20)
  // }

  public sendOptMove(aim: number, pos: cc.Vec2) {
    let opt = new OptMove()
    opt.aim = aim
    opt.px = pos.x
    opt.py = pos.y
    cc.log("[TEST][sendOptMove]x=%d,y=%d,aim=%d", opt.px, opt.py, opt.aim)
    this.sendOptData(OPTTYPE.CLIENT_MOVE, opt, 20)
  }

  private sendOptJump() {
    let menode = this.node.getChildByName(KeyDefine.Key_Node + KeyDefine.Key_Me)
    let opt = new OptJump()
    opt.px = menode.x
    opt.py = menode.y
    cc.log("[TEST][doJump]x=%d,y=%d", opt.px, opt.py)
    this.sendOptData(OPTTYPE.CLIENT_JUMP, opt, 10)
  }

  /**
   * 发送optdata的接口
   * TODO 后面要考虑组包了，如果每次发协议同步多个移动数据的话
   * @param opttype opt类型
   * @param opt 具体opt协议
   * @param len 这里需要估算下opt数据长度，用于创建buffer,可大不可小
   */
  public sendOptData(opttype: number, opt: any, len: number) {
    let optbuf = new ArrayBuffer(len)
    let bio = new biostream()

    // 填充opt
    bio.attach(optbuf)
    opt.write(bio)

    // 填充optdata
    let optdatabuf = new ArrayBuffer(len + 20)
    let optdata = new OptData()
    optdata.numid = MyData.getInstance().data.numid
    optdata.opttype = opttype
    optdata.datalen = bio.getOffset()
    optdata.data = optbuf
    bio.attach(optdatabuf)
    optdata.write(bio)

    // 填充协议
    let req = new msgReqOpt()
    req.index = this.getNowFrameId()
    req.optCnt = 1
    req.optLen = bio.getOffset()
    req.optDatas = optdatabuf
    cc.log("[Game][sendOptData]opttype=%d,id=%d", opttype, req.index)
    NetWork.getInstance().sendMsg(req)
  }

  public checkRTT() {
    var date = new Date();
    let req = new msgCheckRtt()
    req.timestr = date.getTime().toString()
    NetWork.getInstance().sendMsg(req)
  }

  /************************************************************************************************************
   ************************************************************************************************************
   ************************************************************************************************************
   ************************************************************************************************************
   ************************************************************************************************************
   接收网络消息接口
   */

  public onMsg(head: ProtocolHead, message: any): boolean {
    let ret = false
    switch (head.xyid) {
      case XYID.NTF_OPT:
        ret = this.onNtfOpt(head, message)
        break
      case XYID.NTF_GAMEEND:
        ret = this.onNtfGameEnd(head, message)
        break
      case XYID.CHECK_RTT:
        ret = this.onCheckRTT(head, message)
        break
    }
    return ret
  }

  private onNtfOpt(head: ProtocolHead, message: any): boolean {
    let bio = new biostream()
    bio.attach(message)
    let res = new msgNtfOpt()
    res.read(bio)
    cc.log("[Game][onNtfOpt]flag=%d,framecnt=%d,framelen=%d", res.flag, res.frameCnt, res.frameLen)

    let bioframe = new biostream()
    bioframe.attach(res.frameDatas)
    for (let i = 0; i < res.frameCnt; i++) {
      let frame = new FrameData()
      frame.read(bioframe)
      cc.log("[Game][onNtfOpt]index=%d,optcnt=%d", frame.frameindex, frame.optcnt)
      for (let j = 0; j < frame.optcnt; j++) {
        cc.log("[Game][onNtfOpt]numid=%d,opttype=%d,optlen=%d",
          frame.optdatas[j].numid, frame.optdatas[j].opttype, frame.optdatas[j].datalen)
        switch (frame.optdatas[j].opttype) {
          case OPTTYPE.GAME_START:
            this.recvOptGameStart(frame.optdatas[j])
            break
          case OPTTYPE.CLIENT_MOVE:
            this.recvOptMove(frame.optdatas[j])
            break
          case OPTTYPE.CLIENT_JUMP:
            this.recvOptJump(frame.optdatas[j])
            break
          case OPTTYPE.NEW_STAR:
            this.recvOptNewStar(frame.optdatas[j])
            break
          case OPTTYPE.CATCH_STAR:
            this.recvOptCatchStar(frame.optdatas[j])
            break
          // case OPTTYPE.CLIENT_POSITION:
          //   this.recvOptPosition(frame.optdatas[j])
          //   break
          case OPTTYPE.DEL_STAR:
            this.recvOptDelStar(frame.optdatas[j])
            break
        }
      }
    }
    return false
  }

  private recvOptGameStart(optdata: OptData) {
    let bio = new biostream()
    bio.attach(optdata.data)
    let opt = new OptGameStart()
    opt.read(bio)

    for (let i = 0; i < opt.playerCnt; i++) {
      cc.log("[Game][recvOptGameStart]numid=%d,who=%d,nickname=%s,gold=%d,px=%d,py=%d",
        opt.players[i].numid, opt.players[i].who, opt.players[i].nickname,
        opt.players[i].gold, opt.players[i].px, opt.players[i].py)

      // 初始化数据
      if (MyData.getInstance().data.numid == opt.players[i].numid) {

        MyData.getInstance().data.px = opt.players[i].px
        MyData.getInstance().data.py = opt.players[i].py

        let me = cc.instantiate(this.playerPrefab)
        me.setPosition(cc.v2(MyData.getInstance().data.px, MyData.getInstance().data.py))
        me.getComponent('Player').init(this, KeyDefine.Key_Me, true)
        this.node.addChild(me, 0, KeyDefine.Key_Node + KeyDefine.Key_Me)

        let score = cc.instantiate(this.scorePrefab)
        score.setPosition(cc.v2(MyData.getInstance().data.px < 0 ? -250 : 250, 180)) // TODO 积分位置写死
        score.getComponent('Score').init(KeyDefine.Key_Me)
        this.node.addChild(score)

        MyData.getInstance().data.addScore(0)
        MyData.getInstance().data.noticeNickname()

      } else {

        EnemyData.getInstance().data.numid = opt.players[i].numid
        EnemyData.getInstance().data.nickname = opt.players[i].nickname
        EnemyData.getInstance().data.px = opt.players[i].px
        EnemyData.getInstance().data.py = opt.players[i].py

        let enemy = cc.instantiate(this.playerPrefab)
        enemy.setPosition(cc.v2(EnemyData.getInstance().data.px, EnemyData.getInstance().data.py))
        enemy.getComponent('Player').init(this, KeyDefine.Key_Enemey, false)
        this.node.addChild(enemy, 1, KeyDefine.Key_Node + KeyDefine.Key_Enemey)

        let score = cc.instantiate(this.scorePrefab)
        score.setPosition(cc.v2(EnemyData.getInstance().data.px < 0 ? -250 : 250, 180)) // TODO 积分位置写死
        score.getComponent('Score').init(KeyDefine.Key_Enemey)
        this.node.addChild(score)

        EnemyData.getInstance().data.addScore(0)
        EnemyData.getInstance().data.noticeNickname()
      }
    }

    this.gamestart = true
  }

  private recvOptMove(optdata: OptData) {
    let bio = new biostream()
    bio.attach(optdata.data)
    let opt = new OptMove()
    opt.read(bio)

    if (optdata.numid == MyData.getInstance().data.numid) {
      // 自己移动不用网络同步
    } else {
      cc.log("[Game][recvMove]:numid=%d,aim=%d,px=%d,py=%d", optdata.numid, opt.aim, opt.px, opt.py)
      EnemyData.getInstance().data.updateAim(opt.aim)
      EnemyData.getInstance().data.updatePos(cc.v2(opt.px, opt.py))
    }
  }

  private recvOptJump(optdata: OptData) {
    let bio = new biostream()
    bio.attach(optdata.data)
    let opt = new OptJump()
    opt.read(bio)

    if (optdata.numid == MyData.getInstance().data.numid) {
      // 自己跳跃不用网络同步
    } else {
      cc.log("[Game][recvJump]:numid=%d,px=%d,py=%d", optdata.numid, opt.px, opt.py)
      EnemyData.getInstance().data.updateJump(cc.v2(opt.px, opt.py))
    }
  }

  private recvOptNewStar(optdata: OptData) {
    let bio = new biostream()
    bio.attach(optdata.data)
    let opt = new OptNewStar()
    opt.read(bio)

    cc.log("[Game][recvNewStar]i=%d,x=%d,y=%d,duration=%d", opt.index, opt.px, opt.py, opt.duration)
    let newStar = cc.instantiate(this.starPrefab)
    newStar.setPosition(cc.v2(opt.px, opt.py))
    newStar.getComponent('Star').init(this, opt.index)
    this.node.addChild(newStar, 1, KeyDefine.Key_Node + KeyDefine.Key_Star + opt.index)
  }

  private recvOptCatchStar(optdata: OptData) {
    let bio = new biostream()
    bio.attach(optdata.data)
    let opt = new OptCatchStar()
    opt.read(bio)

    cc.log("[Game][recvCatchStar]numid=%d,index=%d", optdata.numid, opt.index)

    let starnode = this.node.getChildByName(KeyDefine.Key_Node + KeyDefine.Key_Star + opt.index)
    if (starnode != null) {
      if (optdata.numid == MyData.getInstance().data.numid) {
        MyData.getInstance().data.addScore(GameConfig.Add_Score)
      } else {
        EnemyData.getInstance().data.addScore(GameConfig.Add_Score)
      }
      starnode.getComponent("Star").onPicked()
    } else {
      cc.warn("[Game][recvCatchStar]numid=%d,can not find=%d", optdata.numid, opt.index)
    }

  }

  // private recvOptPosition(optdata: OptData) {
  //   let bio = new biostream()
  //   bio.attach(optdata.data)
  //   let opt = new OptPosition()
  //   opt.read(bio)

  //   if (optdata.numid == MyData.getInstance().data.numid) {
  //     // 自己位置不需要同步
  //   } else {
  //     cc.log("[Game][recvOptPosition]numid=%d,aim=%d,x=%d,y=%d", optdata.numid, opt.aim, opt.px, opt.py)
  //     EnemyData.getInstance().data.updateAim(opt.aim)
  //     EnemyData.getInstance().data.updatePos(cc.v2(opt.px, opt.py))
  //   }
  // }

  private recvOptDelStar(optdata: OptData) {
    let bio = new biostream()
    bio.attach(optdata.data)
    let opt = new OptDelStar()
    opt.read(bio)

    let starnode = this.node.getChildByName(KeyDefine.Key_Node + KeyDefine.Key_Star + opt.index)
    if (starnode != null) {
      starnode.getComponent("Star").onDel()
    } else {
      cc.warn("[Game][recvOptDelStar]can not find=%d", opt.index)
    }
  }

  private onNtfGameEnd(head: ProtocolHead, message: any): boolean {
    cc.log("[Game][onNtfGameEnd]Game end")
    return true
  }

  private onCheckRTT(head: ProtocolHead, message: any): boolean {
    let date = new Date()

    let bio = new biostream()
    bio.attach(message)
    let res = new msgCheckRtt()
    res.read(bio)
    cc.log("[Game][onCheckRTT]sendtime=%s,nowtime=%d", res.timestr, date.getTime())

    let dif = date.getTime() - parseInt(res.timestr, 0)
    this.rtttips.string = "RTT:" + dif.toString()
    return true
  }

}