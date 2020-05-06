package protocol

const HeadLen = 10

const (
	XYID_HEARTBEAT = 1
	XYID_REQ_SERVERREGIST = 2
	XYID_RES_SERVERREGIST = 3
	XYID_REQ_LOGIN = 4
	XYID_RES_LOGIN = 5
	XYID_REQ_PLAYER_LEAVE = 6
	XYID_REQ_REGIST = 7
	XYID_RES_REGIST = 8
	XYID_CHECK_RTT = 9

	XYID_REQ_JOINROOM = 100
	XYID_RES_JOINROOM = 101
	XYID_REQ_OPT = 102
	XYID_NTF_OPT = 103
	XYID_NTF_GAMESTART = 104
	XYID_NTF_MATCHING = 105
	XYID_NTF_GAMEEND = 106
	XYID_REQ_CONNECT = 107
)

type ProtocolInterface interface {
	Read(s *Biostream)
	Write(s *Biostream)
}

type ProtocolHead struct {
	Length uint16
	Xyid   uint32
	Numid  uint32
}
func (this *ProtocolHead) ReadHead(s *Biostream) {
	this.Length = s.ReadUint16()
	this.Xyid = s.ReadUint32()
	this.Numid = s.ReadUint32()
}
func (this *ProtocolHead) WriteHead(s *Biostream, xyid uint32) {
	this.Xyid = xyid
	s.WriteUint16(0) // 这时候还不知道协议长度
	s.WriteUint32(this.Xyid)
	s.WriteUint32(this.Numid)
}
func (this *ProtocolHead) WriteHeadEnd(s *Biostream) {
	// 有没有更好的办法？
	this.Length = uint16(s.GetOffset())
	s.Seek(0)
	s.WriteUint16(this.Length)
	s.Seek(uint32(this.Length))
}

func ProtocolToBuffer(xy ProtocolInterface, buf []byte) uint32 {
	bio := Biostream{}
	bio.Attach(buf, len(buf))
	xy.Write(&bio)
	return bio.GetOffset()
}

func BufferToProtocol(buf []byte, xy ProtocolInterface) {
	bio := Biostream{}
	bio.Attach(buf, len(buf))
	xy.Read(&bio)
}