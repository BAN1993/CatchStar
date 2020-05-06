import { DataPosters } from "../Data/DataPoster";

const { ccclass, property } = cc._decorator;

@ccclass
export class Score extends cc.Component {
  @property(cc.Label)
  private scoreLabel: cc.Label = null
  @property({ type: cc.AudioClip, })
  private scoreAudio: cc.AudioClip = null

  private keyName: string = ""

  private nickname: string = "nickname"
  private score: number = 0

  protected onLoad() {
    this.scoreLabel.string = this.nickname + ":" + this.score.toString()
  }

  public init(name: string) {
    this.keyName = name
    DataPosters.getInstance().regist(this.keyName + "_score", this)
    DataPosters.getInstance().regist(this.keyName + "_nickname", this)
  }

  public onDataUpdate(key: string, message: any) {
    if (key == this.keyName + "_score") {
      this.scoreLabel.string = this.nickname + ":" + message.toString()
      cc.audioEngine.play(this.scoreAudio as any, false, 1)
    } else if (key == this.keyName + "_nickname") {
      this.nickname = message
      this.scoreLabel.string = this.nickname + ":" + this.score.toString()
    }
  }
}
