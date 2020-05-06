import { biostream } from "./biostream"



export class ProtocolHead {
  public static HeadLen: number = 10

  public length: number = 0
  public xyid: number = 0
  public numid: number = 0

  constructor(xyid: number) {
    this.xyid = xyid
  }

  public read(bio: biostream) {
    this.length = bio.readUint16()
    this.xyid = bio.readUint32()
    this.numid = bio.readUint32()
  }

  public write(bio: biostream) {
    bio.writeUint16(this.length)
    bio.writeUint32(this.xyid)
    bio.writeUint32(this.numid)
  }
}