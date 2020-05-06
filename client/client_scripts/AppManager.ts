import { NetWork } from "./Net/NetWork"
import { NetConfig } from "./NetConfig"

const { ccclass, property } = cc._decorator

@ccclass
export class AppManager extends cc.Component {

  protected onLoad() {

    cc.view.setOrientation(cc.macro.ORIENTATION_PORTRAIT)

    if (!NetWork.getInstance().isConnect()) {
      cc.log("[AppManager]try connect")
      let ret: boolean = NetWork.getInstance().toConnect(NetConfig.serverHost)
      if (!ret) {
        cc.error("[AppManager]connect to %s error", NetConfig.serverHost)
      }
    }
  }

}
