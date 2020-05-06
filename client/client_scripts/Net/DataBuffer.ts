export class DataBuffer {
  private buffer: ArrayBuffer = null
  private length: number = 0
  private index_begin: number = 0
  private index_end: number = 0

  constructor(len: number) {
    this.buffer = new ArrayBuffer(len)
    this.length = len
    this.index_begin = 0
    this.index_end = 0
  }

  public reset() {
    this.index_begin = 0
    this.index_end = 0
  }

  /**
   * 获取缓存区数据长度
   */
  public getLen(): number {
    return this.index_end - this.index_begin
  }

  /**
   * 往缓存添加数据
   */
  public add(src: ArrayBuffer): boolean {
    if (!this.buffer || !src)
      return false

    this.moveBegin()

    if (this.index_end + src.byteLength > this.length)
      return false

    let dataview = new DataView(this.buffer, this.index_end)
    let srcview = new DataView(src)
    for (let i = 0; i < srcview.byteLength; i++) {
      let t = srcview.getUint8(i)
      dataview.setUint8(this.index_end, t)
      this.index_end++
    }
    return true
  }

  /**
   * 消耗缓存数据
   */
  public consume(len: number): ArrayBuffer {
    if (this.index_end - this.index_begin < len)
      return null

    let ret = this.buffer.slice(this.index_begin, this.index_begin + len)
    this.index_begin += len
    return ret
  }

  /**
   * 直接消耗,不返回数据
   */
  public consumePass(len: number) {
    if (this.index_end - this.index_begin < len)
      return
    this.index_begin += len
  }

  /**
   * 获取缓存数据,但是不消耗
   */
  public watch(len: number): ArrayBuffer {
    if (this.index_end - this.index_begin < len)
      return null
    let ret = this.buffer.slice(this.index_begin, this.index_begin + len)
    return ret
  }

  /**
   * 将缓存区有效数据往头部移动
   */
  private moveBegin() {
    if (this.index_begin <= 0)
      return

    if (this.index_begin >= this.index_end) {
      this.index_begin = 0
      this.index_end = 0
      return
    }

    let len = this.index_end - this.index_begin
    let olds = new DataView(this.buffer, this.index_begin, len)
    let news = new DataView(this.buffer, 0, len)
    for (let i = this.index_begin; i < len; i++) {
      let t = olds.getUint8(i)
      news.setUint8(i, t)
    }
    this.index_begin = 0
    this.index_end = len
  }
}
