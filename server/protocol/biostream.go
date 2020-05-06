package protocol

import "encoding/binary"

/*
	1.默认小端
*/

type Biostream struct {
	buffer []byte
	offset uint32
	bufferlen uint32
}

func (this *Biostream) Attach(buffer []byte, len int) {
	this.buffer = buffer
	this.offset = 0
	this.bufferlen = uint32(len)
}

func (this *Biostream) GetOffset() uint32 {
	return this.offset
}

func (this *Biostream) GetBuffer() []byte {
	return this.buffer
}

func (this *Biostream) Seek(index uint32) {
	this.offset = index
}

func (this *Biostream) WriteByte(b byte) {
	if this.offset + 1 > this.bufferlen	{
		panic("overflow")
	}
	this.buffer[this.offset] = b
	this.offset += 1
}

func (this *Biostream) WriteUint8(n uint8) {
	this.WriteByte(byte(n))
}

func (this *Biostream) WriteInt8(n int8) {
	this.WriteByte(byte(n))
}

func (this *Biostream) WriteUint16(n uint16) {
	if this.offset + 2 > this.bufferlen	{
		panic("overflow")
	}
	binary.LittleEndian.PutUint16(
		this.buffer[this.offset : this.offset + 2], n)
	this.offset += 2
}

func (this *Biostream) WriteInt16(n int16) {
	this.WriteUint16(uint16(n))
}

func (this *Biostream) WriteUint32(n uint32) {
	if this.offset + 4 > this.bufferlen	{
		panic("overflow")
	}
	binary.LittleEndian.PutUint32(
		this.buffer[this.offset : this.offset + 4], n)
	this.offset += 4
}

func (this *Biostream) WriteInt32(n int32) {
	this.WriteUint32(uint32(n))
}

func (this *Biostream) WriteUint64(n uint64) {
	if this.offset + 8 > this.bufferlen	{
		panic("overflow")
	}
	binary.LittleEndian.PutUint64(
		this.buffer[this.offset : this.offset + 8], n)
	this.offset += 8
}

func (this *Biostream) WriteInt64(n int64) {
	this.WriteUint64(uint64(n))
}

func (this *Biostream) WriteString(s string) {
	var slen uint32	= uint32(len(s))
	if slen < 0xff {
		this.WriteUint8(uint8(slen))
	} else if slen < 0xfffe {
		this.WriteUint8(0xff)
		this.WriteUint16(uint16(slen))
	} else {
		this.WriteUint8(0xff)
		this.WriteUint16(0xffff)
		this.WriteUint32(slen)
	}
	var i uint32
	for i=0; i<slen; i=i+1 {
		this.buffer[this.offset+i] = s[i]
	}
	this.offset += slen
}

func (this *Biostream) WriteBytes(buf []byte, blen uint32) {
	// 直接写,不附加长度
	if this.offset + blen > this.bufferlen {
		panic("overflow")
	}
	if blen > uint32(len(buf)) {
		panic("write bytes overflow")
	}
	var i uint32 = 0
	for i=0; i<blen; i++ {
		this.buffer[this.offset+i] = buf[i]
	}
	this.offset += blen
}

func (this *Biostream) ReadByte() byte {
	if this.offset + 1 > this.bufferlen	{
		panic("overflow")
	}
	ret := this.buffer[this.offset]
	this.offset += 1
	return ret
}

func (this *Biostream) ReadUint8() uint8 {
	return uint8(this.ReadByte())
}

func (this *Biostream) ReadInt8() int8 {
	return int8(this.ReadByte())
}

func (this *Biostream) ReadUint16() uint16 {
	if this.offset + 2 > this.bufferlen	{
		panic("overflow")
	}
	ret := binary.LittleEndian.Uint16(
		this.buffer[this.offset : this.offset + 2])
	this.offset += 2
	return ret
}

func (this *Biostream) ReadInt16() int16 {
	return int16(this.ReadUint16())
}

func (this *Biostream) ReadUint32() uint32 {
	if this.offset + 4 > this.bufferlen	{
		panic("overflow")
	}
	ret := binary.LittleEndian.Uint32(
		this.buffer[this.offset : this.offset + 4])
	this.offset += 4
	return ret
}

func (this *Biostream) ReadInt32() int32 {
	return int32(this.ReadUint32())
}

func (this *Biostream) ReadUint64() uint64 {
	if this.offset + 8 > this.bufferlen	{
		panic("overflow")
	}
	ret := binary.LittleEndian.Uint64(
		this.buffer[this.offset : this.offset + 8])
	this.offset += 8
	return ret
}

func (this *Biostream) ReadInt64() int64 {
	return int64(this.ReadUint64())
}

func (this *Biostream) ReadString() string {
	blen := this.ReadUint8()
	if blen < 0xff {
		return string(this.ReadBytes(uint32(blen)))
	}

	wlen := this.ReadUint16()
	if wlen < 0xfffe {
		return string(this.ReadBytes(uint32(wlen)))
	}

	dwlen := this.ReadUint32()
	return string(this.ReadBytes(dwlen))
}

func (this *Biostream) ReadBytes(blen uint32) []byte {
	if this.offset + blen > this.bufferlen	{
		panic("overflow")
	}
	ret := this.buffer[this.offset : this.offset + blen]
	this.offset += blen
	return ret
}
