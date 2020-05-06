import { msgResLogin } from "../Protocol/ProtocolMsg"
import { DataPosters } from "./DataPoster"

export class UserData {
  public numid: number = 0
  public nickname: string = ""
  public who: number = 0
  public score: number = 0
  public aim: number = 0
  public px: number = 0
  public py: number = 0

  private keyname: string = ""

  constructor(keyname: string) {
    this.keyname = keyname
  }

  public update(msg: msgResLogin) {
    this.numid = msg.numidxy
    this.nickname = msg.nickname
  }

  public updateAim(aim: number) {
    this.aim = aim
    DataPosters.getInstance().toUpdate(this.keyname + "_aim", aim)
  }

  public updatePos(pos: cc.Vec2) {
    this.px = pos.x
    this.py = pos.y
    DataPosters.getInstance().toUpdate(this.keyname + "_pos", pos)
  }

  public updateJump(pos: cc.Vec2) {
    this.px = pos.x
    this.py = pos.y
    DataPosters.getInstance().toUpdate(this.keyname + "_jump", pos)
  }

  public updateNickname(name: string) {
    this.nickname = name
    this.noticeNickname()
  }

  public noticeNickname() {
    DataPosters.getInstance().toUpdate(this.keyname + "_nickname", this.nickname)
  }

  public addScore(num: number) {
    this.score += num
    DataPosters.getInstance().toUpdate(this.keyname + "_score", this.score)
  }
}