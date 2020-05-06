/**
 * 流模块,提供自定义的序列化和反序列化的接口
 * TODO 写字符串没有提供长度限制
 */

export class biostream {
  public data: ArrayBuffer = null
  public length: number = 0
  public offset: number = 0

  public attach(buf: ArrayBuffer) {
    this.data = buf
    this.length = this.data.byteLength
    this.offset = 0
  }

  public setOffset(offset: number) {
    if (offset > this.length) {
      throw "Can not set offset!len=" + this.length + ",offset=" + offset
    }
    this.offset = offset
  }

  public getOffset(): number {
    return this.offset
  }

  public readUint8(): number {
    let view = new DataView(this.data, this.offset, 1)
    this.offset += 1
    return view.getUint8(0)
  }

  public readUint16(): number {
    let view = new DataView(this.data, this.offset, 2)
    this.offset += 2
    return view.getUint16(0, true)
  }

  public readUint32(): number {
    let view = new DataView(this.data, this.offset, 4)
    this.offset += 4
    return view.getUint32(0, true)
  }

  public readInt32(): number {
    let view = new DataView(this.data, this.offset, 4)
    this.offset += 4
    return view.getInt32(0, true)
  }

  public readString(): string {
    let blen = this.readUint8()
    if (blen < 255) {
      return this.readBytes(blen)
    }
    let wlen = this.readUint16()
    if (wlen < 65534) {
      return this.readBytes(wlen)
    }
    let dwlen = this.readUint32()
    return this.readBytes(dwlen)
  }

  public readBytes(len: number): string {
    let str = String.fromCharCode.apply(null,
      new Uint8Array(this.data, this.offset, len))
    this.offset += len
    return str
  }

  public readArrayBuffer(len: number): ArrayBuffer {
    let ret = this.data.slice(this.offset, this.offset + len)
    this.offset += len
    return ret
  }

  public writeUint8(num: number): boolean {
    if (this.offset + 1 > this.length) {
      throw "Buffer overflow! offset=" + this.offset + ",len=" + this.length
    }
    let view = new DataView(this.data, this.offset, 1)
    view.setUint8(0, num)
    this.offset += 1
    return true
  }

  public writeUint16(num: number): boolean {
    if (this.offset + 2 > this.length) {
      throw "Buffer overflow! offset=" + this.offset + ",len=" + this.length
    }
    let view = new DataView(this.data, this.offset, 2)
    view.setUint16(0, num, true)
    this.offset += 2
    return true
  }

  public writeUint32(num: number): boolean {
    if (this.offset + 4 > this.length) {
      throw "Buffer overflow! offset=" + this.offset + ",len=" + this.length
    }
    let view = new DataView(this.data, this.offset, 4)
    view.setUint32(0, num, true)
    this.offset += 4
    return true
  }

  public writeInt32(num: number): boolean {
    if (this.offset + 4 > this.length) {
      throw "Buffer overflow! offset=" + this.offset + ",len=" + this.length
    }
    let view = new DataView(this.data, this.offset, 4)
    view.setInt32(0, num, true)
    this.offset += 4
    return true
  }

  public writeString(str: string): boolean {
    let len = str.length
    if (len < 255) {
      this.writeUint8(len)
    } else if (len < 65534) {
      this.writeUint8(256)
      this.writeUint16(len)
    } else {
      this.writeUint8(256)
      this.writeUint16(65535)
      this.writeUint32(len)
    }
    return this.writeBytes(str)
  }

  public writeBytes(str: string): boolean {
    if (this.offset + str.length > this.length) {
      throw "Buffer overflow! offset=" + this.offset + ",strlen=" + str.length + ",len=" + this.length
    }
    let view = new DataView(this.data, this.offset, str.length)
    for (let i = 0; i < str.length; i++) {
      view.setUint8(i, str.charCodeAt(i))
      this.offset += 1
    }
    return true
  }

  public writeArrayBuffer(array: ArrayBuffer, len: number): boolean {
    if (array.byteLength < len) {
      throw "Array is not enough!bytelen=" + array.byteLength + ",trylen=" + len
    }
    if (this.offset + len > this.length) {
      throw "Buffer overflow! offset=" + this.offset + ",arraylen=" + len + ",len=" + this.length
    }
    let view = new DataView(this.data, this.offset, len)
    let srcview = new DataView(array)
    for (let i = 0; i < len; i++) {
      let t = srcview.getUint8(i)
      view.setUint8(i, t)
      this.offset += 1
    }
    return true
  }
}