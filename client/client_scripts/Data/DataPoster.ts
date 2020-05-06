import { UserData } from "./UserData";
import { NetWork } from "../Net/NetWork";
import { ProtocolHead } from "../Protocol/ProtocolHead";

const { ccclass, property } = cc._decorator;

@ccclass
export class DataPosters {
  private static instance: DataPosters = null
  private delegates: Map<string, any[]> = new Map()
  //private delegates: { [key: string]: any[] } = {}

  public static getInstance(): DataPosters {
    if (!this.instance) {
      this.instance = new DataPosters()
      this.instance.init()
    }
    return this.instance
  }

  public init() {
    NetWork.getInstance().regist(this)
  }

  /**
   * 添加到监听队列
   * 需要实现：
   * public onDataUpdate(key: string, message: any)
   * @param key 监听key
   * @param delegate this
   */
  public regist(key: string, delegate: any) {
    cc.log("[DataPosters][regist]key=%s,delegate=", key, delegate)
    if (this.delegates.has(key)) {
      this.delegates.get(key).push(delegate)
    }
    else {
      let l: any[] = []
      this.delegates.set(key, l)
      this.delegates.get(key).push(delegate)
    }
  }

  public unregist(key: string, delegate: any) {
    cc.log("[DataPosters][unregist]key=%s,delegate=", key, delegate)
    if (this.delegates.has(key)) {
      for (let i = 0; i < this.delegates[key].length; i++) {
        if (this.delegates.get(key)[i] == delegate) {
          this.delegates.get(key).splice(i, 1)
          return true
        }
      }
    }
  }

  /**
   * 协议触发
   */
  public onMsg(head: ProtocolHead, message: any): boolean {
    return false
  }

  /**
   * 主动触发
   */
  public toUpdate(key: string, message: any): boolean {
    return this.post(key, message)
  }

  private post(key: string, message: any): boolean {
    cc.log("[DataPoster][post]key=%s,message=", key, message)
    if (this.delegates.has(key)) {
      let list: any[] = this.delegates.get(key)
      for (let i = 0; i < list.length; i++) {
        if (list[i] && typeof (list[i].onDataUpdate) == "function") {
          list[i].onDataUpdate(key, message)
        } else {
          cc.error("[DataPoster][post]had no (onDataUpdate) function:", typeof (list[i].onDataUpdate))
        }
      }
      return true
    } else {
      cc.log("[DataPoster][post]not have key=%s", key)
    }
    return false
  }
}