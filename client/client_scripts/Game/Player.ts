/* 目前同步方式
 * 自己:
 *    根据操作实时绘制,并每帧发送位置数据
 * 敌方:
 *    根据OptMove同步aim和x,根据aim预测下一帧x坐标
 *    根据OptJump同步y  
 */

import { DataPosters } from "../Data/DataPoster"
import { ActionTag, GameConfig, MoveType, KeyDefine } from "../SomeDefine"
import { Game } from "./Game"

const { ccclass, property } = cc._decorator

@ccclass
export class Player extends cc.Component {
  @property(cc.Integer)
  private jumpHeight: number = 0
  @property(cc.Integer)
  private jumpDuration: number = 0
  @property(cc.Integer)
  private moveSpeed: number = 0
  @property({ type: cc.AudioClip, })
  private jumpAudio: cc.AudioClip = null
  @property(cc.Label)
  private nicknameLabel: cc.Label = null

  private game: Game = null
  private canJump: boolean = true // 能否跳跃
  private standAction: cc.Action = null // 站立动作
  private moveAction: cc.Action = null // 移动动作
  private keyName: string = "" // node键值
  private jumpAction: cc.Action = null // 跳跃动作
  private aim: number = MoveType.Stop // 方向-自己控制用
  private lastMyFrameId: number = 0 // 上一次自己同步的帧号

  private contorlByMyself: boolean = true // 由自己控制或协议控制
  private msgPX: number = 0 // 对端坐标x
  private msgPY: number = 0 // 对端坐标y

  protected onLoad() {
    this.standAction = this.getStandAction()
    this.node.runAction(this.standAction)

    this.moveAction = this.getMoveAction()

    this.jumpAction = this.getJumpAction()
    this.canJump = true
  }

  protected update(dt: number) {
    if (this.contorlByMyself) {
      this.updatePositionMyself(dt)
    } else {
      this.updatePositionMsg(dt)
    }
    if (this.keyName == KeyDefine.Key_Me) {
      // 告知自己位置
      if (this.lastMyFrameId != this.game.getNowFrameId()) {
        this.lastMyFrameId = this.game.getNowFrameId()
        this.game.sendOptMove(this.aim, cc.v2(this.node.x, this.node.y))
      }
    }
  }

  /************************************************************************************************************
   ************************************************************************************************************
   ************************************************************************************************************
   ************************************************************************************************************
   ************************************************************************************************************
   同步
   */

  /**
   * 由自己控制坐标
   */
  private updatePositionMyself(dt: number) {
    let aimfix = 0
    switch (this.aim) {
      case MoveType.Stop:
        aimfix = 0
        break
      case MoveType.Left:
        aimfix = -1
        break
      case MoveType.Right:
        aimfix = 1
        break
    }
    this.node.x += this.moveSpeed * aimfix * dt

    // 不允许超出屏幕
    if (Math.abs(this.node.x) > cc.winSize.width / 2) {
      if (this.node.x > 0)
        this.node.x = cc.winSize.width / 2
      else
        this.node.x = -cc.winSize.width / 2
    }

  }

  /**
   * 由网络控制坐标
   */
  private updatePositionMsg(dt: number) {
    this.updatePositionMyself(dt)
  }

  /************************************************************************************************************
   ************************************************************************************************************
   ************************************************************************************************************
   ************************************************************************************************************
   ************************************************************************************************************
   自定义接口
   */

  public init(game: Game, name: string, myself: boolean) {
    this.game = game
    this.contorlByMyself = myself
    this.registData(name)
  }

  private getStandAction() {
    let stretch = cc.scaleTo(this.jumpDuration * 3, 1, 1.03)
    let shrink = cc.scaleTo(this.jumpDuration * 3, 1, 0.97)
    let extraSequence = cc.sequence(stretch, shrink)
    return cc.repeatForever(extraSequence)
  }

  private getMoveAction() {
    let stretch = cc.scaleTo(this.jumpDuration * 1, 0.95, 1.05)
    let shrink = cc.scaleTo(this.jumpDuration * 1, 1.05, 0.95)
    let extraSequence = cc.sequence(stretch, shrink)
    return cc.repeatForever(extraSequence)
  }

  private getJumpAction() {
    // 跳跃
    let jumpUp = cc.moveBy(this.jumpDuration, cc.v2(0, this.jumpHeight)).easing(cc.easeCubicActionOut())
    let jumpDown = cc.moveBy(this.jumpDuration, cc.v2(0, -this.jumpHeight)).easing(cc.easeCubicActionIn())
    let setjumpflag = cc.callFunc(this.setCanJump, this, true)
    let jumpSequence = cc.sequence(jumpUp, jumpDown, setjumpflag)
    // 弹性
    let stretch = cc.scaleTo(this.jumpDuration, 0.85, 1.15).easing(cc.easeCubicActionOut())
    let shrink = cc.scaleTo(this.jumpDuration, 1.15, 0.85).easing(cc.easeCubicActionIn())
    let extraSequence = cc.sequence(stretch, shrink)
    // 跳跃声音
    let callback = cc.callFunc(this.playJumpSound, this)
    // 组合
    let spawn = cc.spawn(jumpSequence, extraSequence, callback)
    return spawn
  }

  private playJumpSound() {
    // 调用声音引擎播放声音
    cc.audioEngine.play(this.jumpAudio as any, false, 1)
  }

  private setCanJump(node: any, flag: boolean) {
    this.canJump = flag
    cc.log("[setCanJump][%s]flag=", this.keyName, this.canJump)
  }

  public getCanJump(): boolean {
    cc.log("[getCanJump][%s]flag=", this.keyName, this.canJump)
    return this.canJump
  }

  /************************************************************************************************************
   ************************************************************************************************************
   ************************************************************************************************************
   ************************************************************************************************************
   ************************************************************************************************************
   数据更新
   */

  private registData(name: string) {
    this.keyName = name
    DataPosters.getInstance().regist(this.keyName + "_aim", this)
    DataPosters.getInstance().regist(this.keyName + "_pos", this)
    DataPosters.getInstance().regist(this.keyName + "_jump", this)
    DataPosters.getInstance().regist(this.keyName + "_nickname", this)
  }

  public onDataUpdate(key: string, message: any) {
    cc.log("[Player][%s][onDataUpdate]key=%s,message=", this.keyName, key, message)
    if (key == this.keyName + "_aim") {

      let lastaim = this.aim
      this.aim = message
      if (this.aim != MoveType.Stop) {
        if (lastaim == MoveType.Stop) {
          this.node.runAction(this.moveAction)
        }
      } else {
        if (lastaim != MoveType.Stop) {
          this.node.stopAction(this.moveAction)
        }
      }

    } else if (key == this.keyName + "_pos") {

      if (!this.contorlByMyself) {
        let disx = Math.abs(message.x - this.node.x)
        if (disx >= GameConfig.NetMoveFixDis) { // 相差较小就算了，不然抖动太明显
          cc.log("[onDataUpdate._pos][%s]now(%d,%d) to(%d,%d)", this.keyName, this.node.x, this.node.y, message.x, message.y)

          // 先停掉上一次fix
          if (this.node.getActionByTag(ActionTag.FixPosition)) {
            this.node.stopActionByTag(ActionTag.FixPosition)
          }
          // 执行fix
          let fixmove = cc.moveTo(GameConfig.NetFixDuration, cc.v2(message.x, this.node.y))
          fixmove.setTag(ActionTag.FixPosition)
          this.node.runAction(fixmove)

        }
      }

    } else if (key == this.keyName + "_jump") {

      cc.log("[onDataUpdate][%s]jump,x=%d,y=%d,nowx=%d,nowy=%d", this.keyName, message.x, message.y, this.node.x, this.node.y)
      // 先停掉上一次的跳跃
      // 不能直接用jumpAction，因为可能还要加个移动动作
      let last = this.node.getActionByTag(ActionTag.FixJump)
      if (last) {
        cc.log("[onDataUpdate][%s]stop last jump", this.keyName)
        this.node.stopActionByTag(ActionTag.FixJump)
        this.setCanJump(null, true)
      }
      // 同步位置
      let action = null
      let disy = Math.abs(message.y - this.node.y)
      if (disy != 0) {
        let fixmove = cc.moveTo(GameConfig.NetFixDuration, cc.v2(message.x, message.y))
        action = cc.sequence(fixmove, this.jumpAction as any)
      } else {
        action = this.jumpAction
      }
      this.setCanJump(this, false)
      action.setTag(ActionTag.FixJump)
      this.node.runAction(action)

    } else if (key == this.keyName + "_nickname") {

      this.nicknameLabel.string = message

    }
  }
}
