import { UserData } from "./UserData";
import { KeyDefine } from "../SomeDefine";

const { ccclass, property } = cc._decorator;

@ccclass
export class MyData {
  public data: UserData = new UserData(KeyDefine.Key_Me)
  private static instance: MyData = null

  public account: string = ""
  public password: string = ""

  public static getInstance(): MyData {
    if (!this.instance) {
      this.instance = new MyData()
    }
    return this.instance
  }
}