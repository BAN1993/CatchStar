import { UserData } from "./UserData";
import { KeyDefine } from "../SomeDefine";

const { ccclass, property } = cc._decorator;

@ccclass
export class EnemyData {
  public data: UserData = new UserData(KeyDefine.Key_Enemey)
  private static instance: EnemyData = null

  public static getInstance(): EnemyData {
    if (!this.instance) {
      this.instance = new EnemyData()
    }
    return this.instance
  }


}