import { DataBuffer } from "./DataBuffer"
import { ProtocolHead } from "../Protocol/ProtocolHead"
import { biostream } from "../Protocol/biostream"
import { MyData } from "../Data/MyData"
import { NetConfig } from "../NetConfig"

/**
 * 网络模块
 * TODO 重连机制
 */

const { ccclass, property } = cc._decorator

export enum NetWorkState {
  NW_STATE_NONE,
  NW_STATE_CONNECTING,
  NW_STATE_CONNECTED,
  NW_STATE_CLOSED,
}

@ccclass
export class NetWork {
  private static instance: NetWork = null

  private destHost: string = ""
  private webSocket: WebSocket = null
  private netWorkState: NetWorkState = NetWorkState.NW_STATE_NONE
  private delegates: any[] = []
  private dataBuffer: DataBuffer = null
  private tempBuf: ArrayBuffer = null

  public static getInstance(): NetWork {
    if (!this.instance) {
      this.instance = new NetWork()
      this.instance.init()
    }
    return this.instance
  }

  public toConnect(dst: string): boolean {
    this.destHost = dst
    return this.doConnect()
  }

  public toClose(): void {
    this.destHost = ""
    if (this.webSocket) {
      this.webSocket.onopen = () => { }
      this.webSocket.onclose = () => { }
      this.webSocket.onerror = () => { }
      this.webSocket.onmessage = () => { }
      this.webSocket.close()
    }
    this.onClose(null)
  }

  /**
   * 添加到监听队列
   * 需要实现:
   * public onMsg(msg: any):boolean{....}
   * 如果确定这个消息只在你这处理，就return true;否则return false
   * @param delegate 需要添加监听的实例
   */
  public regist(delegate: any): void {
    this.delegates.push(delegate)
  }

  /**
   * 移除监听
   * 移除成功返回true,false说明队列中不存在
   * @param delegate 需要移除的实例
   */
  public unregist(delegate: any): boolean {
    for (let i = 0; i < this.delegates.length; i++) {
      if (this.delegates[i] == delegate) {
        this.delegates.splice(i, 1)
        return true
      }
    }
    return false
  }

  public getNetWorkState(): NetWorkState {
    return this.netWorkState
  }

  public isConnect(): boolean {
    return this.getNetWorkState() == NetWorkState.NW_STATE_CONNECTED
  }

  private init(): void {
    this.netWorkState = NetWorkState.NW_STATE_NONE
    this.dataBuffer = new DataBuffer(NetConfig.MAX_BUFFER_SIZE)
    this.tempBuf = new ArrayBuffer(NetConfig.MAX_BUFFER_SIZE)
  }

  /**
   * 真正处理连接的地方
   */
  public doConnect(): boolean {
    if (this.getNetWorkState() == NetWorkState.NW_STATE_CONNECTED
      || this.getNetWorkState() == NetWorkState.NW_STATE_CONNECTING) {
      cc.error("[NetWork]already connect,state=" + this.getNetWorkState())
      return false
    }
    if (!this.destHost || this.destHost.length < 0) {
      cc.error("[NetWork]dest Host is null or error")
      return false
    }
    cc.log("[NetWork]connect to:" + this.destHost)
    this.netWorkState = NetWorkState.NW_STATE_CONNECTING
    this.webSocket = new WebSocket(this.destHost)
    this.webSocket.binaryType = "arraybuffer";//设置数据类型
    this.webSocket.onopen = this.onOpen.bind(this)
    this.webSocket.onclose = this.onClose.bind(this)
    this.webSocket.onerror = this.onError.bind(this)
    this.webSocket.onmessage = this.onMessage.bind(this)
    return true
  }

  private onOpen(ev): void {
    cc.log("[NetWork]open")
    this.netWorkState = NetWorkState.NW_STATE_CONNECTED
  }

  private onClose(ev): void {
    cc.error("[NetWork]close")
    this.netWorkState = NetWorkState.NW_STATE_CLOSED
  }

  private onError(ev): void {
    cc.error("[NetWork]error")
  }

  private onMessage(ev): void {
    this.dataBuffer.add(ev.data)
    let headBuf: ArrayBuffer = this.dataBuffer.watch(ProtocolHead.HeadLen)
    let bio = new biostream()
    while (headBuf != null) {

      // 获取协议头
      let head = new ProtocolHead(0)
      bio.attach(headBuf)
      head.read(bio)

      // 简单验证包头有效性
      if (head.xyid <= 0 || head.xyid >= 65535 ||
        head.length < 0 || head.length > NetConfig.MAX_BUFFER_SIZE) {
        cc.error("[NetWork][onMessage]Recv a error head,xyid=%d,length=%d", head.xyid, head.length)
        this.dataBuffer.reset()
        return
      }
      if (this.dataBuffer.getLen() < head.length) {
        cc.log("[NetWork][onMessage][TEST]get a small buffer wait,buflen=%d,headlen=%d", this.dataBuffer.getLen(), head.length)
        return
      }
      this.dataBuffer.consumePass(ProtocolHead.HeadLen)

      // 获取协议体
      let bodyBuf: ArrayBuffer = this.dataBuffer.consume(head.length - ProtocolHead.HeadLen)
      if (bodyBuf == null) {
        // 说明数据还没接收完
        return
      }

      cc.log("[NetWork][onMessage]recv xyid=%d,delegates=%d", head.xyid, this.delegates.length)
      for (let idx = 0; idx < this.delegates.length; idx++) {
        if (this.delegates[idx]
          && typeof (this.delegates[idx].onMsg) === "function"
          && this.delegates[idx].onMsg(head, bodyBuf)) {
          break
        }
      }

      headBuf = this.dataBuffer.watch(ProtocolHead.HeadLen)
    }
  }

  public sendMsg(msg: any, askid: number = null, callback: Function = null): boolean {
    if (this.isConnect()) {
      // 填充协议体
      let bio = new biostream()
      bio.attach(this.tempBuf)
      bio.setOffset(ProtocolHead.HeadLen)
      msg.write(bio)
      // 填充协议头
      msg.head.length = bio.getOffset()
      msg.head.numid = MyData.getInstance().data.numid
      bio.setOffset(0)
      msg.head.write(bio)
      // 发送数据
      let sendbuf: ArrayBuffer = this.tempBuf.slice(0, msg.head.length)
      this.webSocket.send(sendbuf)
      return true

    } else {
      cc.error("[NetWork][onMessage]send msg error,state=" + this.getNetWorkState())
    }
    return false
  }

}