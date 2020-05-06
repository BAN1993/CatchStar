import { NetWork, NetWorkState } from "../Net/NetWork"
import { biostream } from "../Protocol/biostream"
import { ProtocolHead } from "../Protocol/ProtocolHead"
import { msgReqLogin, msgReqJoinRoom, msgResJoinRoom, msgReqConnect, msgNtfOpt } from "../Protocol/ProtocolMsg"
import { msgResLogin } from "../Protocol/ProtocolMsg"
import { XYID } from "../Protocol/ProtocolXYID"
import { MyData } from "../Data/MyData"

const { ccclass, property } = cc._decorator

@ccclass
export class Button_OK extends cc.Component {
  @property(cc.EditBox)
  private accountBox: cc.Label = null
  @property(cc.EditBox)
  private passwordBox: cc.Label = null
  @property(cc.Label)
  private tips: cc.Label = null

  onLoad() {

    // tips
    this.tips.string = ""

    // 创建按钮事件
    let clickEventHandler = new cc.Component.EventHandler()
    clickEventHandler.target = this.node
    clickEventHandler.component = "Button_OK"
    clickEventHandler.handler = "btnClick"

    let button = this.node.getComponent(cc.Button)
    button.clickEvents.push(clickEventHandler)

    // 注册网络事件
    NetWork.getInstance().regist(this)
  }

  btnClick(event, customEventData) {
    let node = event.target
    let button = node.getComponent(cc.Button)

    if (NetWork.getInstance().getNetWorkState() != NetWorkState.NW_STATE_CONNECTED) {
      this.tips.string = "Can not connect to service!"
      return
    }

    let account = this.accountBox.string
    let pass = this.passwordBox.string

    if (account.length <= 0 || account.length >= 60) {
      this.tips.string = "Invalid Account"
      return
    }
    if (pass.length <= 0 || pass.length >= 60) {
      this.tips.string = "Invaild Password"
      return
    }

    cc.log("[Button_OK]Send reqLogin,account=%s", account)

    let req = new msgReqLogin()
    req.account = account
    req.password = pass
    NetWork.getInstance().sendMsg(req)

    let me = MyData.getInstance()
    me.account = req.account
    me.password = req.password
  }

  public onMsg(head: ProtocolHead, message: any): boolean {
    let ret = false
    switch (head.xyid) {
      case XYID.RES_LOGIN:
        ret = this.onResLogin(head, message)
        break
      case XYID.RES_JOINROOM:
        ret = this.onResJoinRoom(head, message)
        break
      case XYID.NTF_MATCHING:
        ret = true
        this.tips.string = "Matching..."
        break
      case XYID.NTF_GAMESTART:
        ret = this.onNtfGameStart(head, message)
        break
    }
    return ret
  }

  private onResLogin(head: ProtocolHead, message: any): boolean {
    let bio = new biostream()
    bio.attach(message)
    let res = new msgResLogin()
    res.read(bio)

    if (res.flag == msgResLogin.SUCCESS) {
      cc.log("[Button_OK]Login success,try join room")

      MyData.getInstance().data.update(res)

      let req = new msgReqJoinRoom()
      req.gwid = 0
      req.nickname = MyData.getInstance().data.nickname
      NetWork.getInstance().sendMsg(req)
      return true
    }
    cc.error("[Button_OK]Login error,flag=%d", res.flag)
    this.tips.string = "Login error(" + res.flag + ")"
    return true
  }

  private onResJoinRoom(head: ProtocolHead, message: any): boolean {
    let bio = new biostream()
    bio.attach(message)
    let res = new msgResJoinRoom()
    res.read(bio)

    if (res.flag == msgResJoinRoom.SUCCESS) {
      cc.log("[Button_OK]Join room success,now matching...")
      return true
    }
    cc.error("[Button_OK]Join room error,flag=%d", res.flag)
    this.tips.string = "Join error(" + res.flag + ")"
    return true
  }

  private onNtfGameStart(head: ProtocolHead, message: any): boolean {
    // 切场景
    NetWork.getInstance().unregist(this)
    cc.director.loadScene('Game')
    return true
  }

}
