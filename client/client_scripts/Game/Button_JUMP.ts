import { Game } from "./Game";
import { DataPosters } from "../Data/DataPoster";
import { KeyDefine } from "../SomeDefine";

const { ccclass, property } = cc._decorator;

@ccclass
export class Button_JUMP extends cc.Component {
  onLoad() {
    let clickEventHandler = new cc.Component.EventHandler()
    clickEventHandler.target = this.node
    clickEventHandler.component = "Button_JUMP"
    clickEventHandler.handler = "btnClick"

    let button = this.node.getComponent(cc.Button)
    button.clickEvents.push(clickEventHandler)
  }

  btnClick(event, customEventData) {
    DataPosters.getInstance().toUpdate(KeyDefine.Key_CTL + KeyDefine.Key_Me, 1)
  }

}
