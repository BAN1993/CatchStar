import { Game } from "./Game"
import { OptCatchStar, OPTTYPE } from "./Opt"
import { KeyDefine } from "../SomeDefine"

const { ccclass, property } = cc._decorator

@ccclass
export class Star extends cc.Component {
  @property(cc.Integer)
  private pickRadius: number = 0
  @property(cc.Integer)
  private showDownDistance: number = 0
  @property(cc.Integer)
  private showDownDuration: number = 0

  private game: Game = null
  private touchable: boolean = true
  private index: number = 0

  onLoad() {
    this.node.runAction(this.getShowUpAction())
  }

  update(dt: number) {
    if (!this.touchable)
      return
    // 本地逻辑只用检测自己和星星的距离
    if (this.getPlayerDistance(KeyDefine.Key_Me) < this.pickRadius) {
      this.tryPick()
      return
    }
  }

  /************************************************************************************************************
   ************************************************************************************************************
   ************************************************************************************************************
   ************************************************************************************************************
   ************************************************************************************************************
   自定义接口
   */

  public init(game: Game, index: number) {
    this.game = game
    this.index = index
  }

  private getPlayerDistance(keyname: string) {
    let playerPos = this.game.node.getChildByName(KeyDefine.Key_Node + keyname).getPosition()
    let dist = this.node.getPosition().sub(playerPos).mag()
    return dist
  }

  private getPlayerLeft(keyname: string): boolean {
    let playerPos = this.game.node.getChildByName(KeyDefine.Key_Node + keyname).getPosition().x
    let left = this.node.position.x - playerPos
    return left > 0 ? true : false
  }

  private getShowUpAction() {
    this.node.opacity = 0
    let down = cc.moveBy(this.showDownDuration, cc.v2(0, -this.showDownDistance)).easing(cc.easeCubicActionOut())
    let fadein = cc.fadeIn(this.showDownDuration * 2)
    let spawn = cc.spawn(down, fadein)
    return spawn
  }

  private getDestroyAction() {
    // 上升
    let left = this.getPlayerLeft(KeyDefine.Key_Me)
    let movex = this.showDownDistance * 3 * (left ? 1 : -1)
    let up = cc.moveBy(this.showDownDuration, cc.v2(movex, this.showDownDistance * 3)).easing(cc.easeCubicActionOut())
    // 渐出
    let fadeout = cc.fadeOut(this.showDownDuration)
    // 翻转
    let rotate = cc.rotateBy(this.showDownDuration, 360)
    // 缩小
    let small = cc.scaleTo(this.showDownDuration, 0.5, 0.5)
    // 合并
    let spawn = cc.spawn(up, fadeout, rotate, small)
    return spawn
  }

  private getDelAction() {
    // 上升
    //let up = cc.moveBy(this.showDownDuration / 3, cc.v2(0, this.showDownDistance * 3)).easing(cc.easeCubicActionOut())
    // 放大
    let big = cc.scaleTo(this.showDownDuration / 3, 1.5, 1.5).easing(cc.easeCubicActionOut())
    // 渐出
    let fadeout = cc.fadeOut(this.showDownDuration / 3)
    // 缩小
    //let small = cc.scaleTo(this.showDownDuration / 5, 0.5, 0.5)
    // 合并
    let spawn = cc.spawn(big, fadeout)
    return spawn
  }

  /**
   * 真正被抓住的接口，由协议触发
   */
  public onPicked(keyname: string) {
    this.touchable = false
    cc.log("[Star][onPicked]index=%d", this.index)
    // 销毁计时器
    setTimeout(function () { this.node.destroy() }.bind(this), this.showDownDuration * 1000)
    // 销毁动画
    this.node.runAction(this.getDestroyAction())
  }

  public onDel() {
    this.touchable = false
    cc.log("[Star][onDel]index=%d", this.index)
    // 销毁计时器
    setTimeout(function () { this.node.destroy() }.bind(this), this.showDownDuration * 1000)
    // 销毁动画
    this.node.runAction(this.getDelAction())
  }

  /**
   * 尝试抓取接口，由本地调用
   */
  private tryPick() {
    if (!this.touchable)
      return
    this.touchable = false

    let opt = new OptCatchStar()
    opt.index = this.index
    cc.log("[Star][tryPick]index=%d", opt.index)
    this.game.sendOptData(OPTTYPE.CATCH_STAR, opt, 10)
  }

}
