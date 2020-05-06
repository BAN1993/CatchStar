import { biostream } from "./biostream"
import { ProtocolHead } from "./ProtocolHead"
import { XYID } from "./ProtocolXYID"

export class msgReqLogin {
  public head: ProtocolHead = new ProtocolHead(XYID.REQ_LOGIN)
  public account: string = ""
  public password: string = ""

  public write(bio: biostream) {
    bio.writeString(this.account)
    bio.writeString(this.password)
  }
}

export class msgResLogin {

  public static SUCCESS: number = 0

  public head: ProtocolHead = new ProtocolHead(XYID.RES_LOGIN)
  public flag: number = 0
  public numidxy: number = 0
  public nickname: string = ""

  public read(bio: biostream) {
    this.flag = bio.readUint32()
    this.numidxy = bio.readUint32()
    this.nickname = bio.readString()
  }
}

export class msgCheckRtt {

  public head: ProtocolHead = new ProtocolHead(XYID.CHECK_RTT)
  public timestr: string = ""

  public read(bio: biostream) {
    this.timestr = bio.readString()
  }

  public write(bio: biostream) {
    bio.writeString(this.timestr)
  }
}

export class msgReqJoinRoom {
  public head: ProtocolHead = new ProtocolHead(XYID.REQ_JOINROOM)
  public gwid: number = 0
  public nickname: string = ""

  public write(bio: biostream) {
    bio.writeUint32(this.gwid)
    bio.writeString(this.nickname)
  }
}

export class msgResJoinRoom {

  public static SUCCESS: number = 0

  public head: ProtocolHead = new ProtocolHead(XYID.RES_JOINROOM)
  public gwid: number = 0
  public flag: number = 0
  public nickname: string = ""

  public read(bio: biostream) {
    this.gwid = bio.readUint32()
    this.flag = bio.readUint32()
    this.nickname = bio.readString()
  }
}

export class msgReqConnect {
  public head: ProtocolHead = new ProtocolHead(XYID.REQ_CONNECT)

  public write(bio: biostream) {

  }
}

export class msgReqOpt {
  public head: ProtocolHead = new ProtocolHead(XYID.REQ_OPT)
  public index: number = 0
  public optCnt: number = 0
  public optLen: number = 0
  public optDatas: ArrayBuffer = null

  public write(bio: biostream) {
    bio.writeUint32(this.index)
    bio.writeUint32(this.optCnt)
    bio.writeUint32(this.optLen)
    bio.writeArrayBuffer(this.optDatas, this.optLen)
  }
}

export class msgNtfOpt {
  public head: ProtocolHead = new ProtocolHead(XYID.NTF_OPT)
  public flag: number = 0
  public frameCnt: number = 0
  public frameLen: number = 0
  public frameDatas: ArrayBuffer = null

  public read(bio: biostream) {
    this.flag = bio.readUint32()
    this.frameCnt = bio.readUint32()
    this.frameLen = bio.readUint32()
    this.frameDatas = bio.readArrayBuffer(this.frameLen)
  }
}