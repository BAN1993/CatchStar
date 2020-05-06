package protocol

type HeartBeat struct {
	ProtocolInterface
	ProtocolHead
	Timestamp uint32
}
func (this *HeartBeat) Read(s *Biostream) {
	this.ReadHead(s)
	this.Timestamp = s.ReadUint32()
}
func (this *HeartBeat) Write(s *Biostream) {
	this.WriteHead(s, XYID_HEARTBEAT)
	s.WriteUint32(this.Timestamp)
	this.WriteHeadEnd(s)
}

type ReqServerRegist struct {
	ProtocolInterface
	ProtocolHead
	Serverid uint32
	Servertype uint32
}
func (this *ReqServerRegist) Read(s *Biostream) {
	this.ReadHead(s)
	this.Serverid = s.ReadUint32()
	this.Servertype = s.ReadUint32()
}
func (this *ReqServerRegist) Write(s *Biostream) {
	this.WriteHead(s, XYID_REQ_SERVERREGIST)
	s.WriteUint32(this.Serverid)
	s.WriteUint32(this.Servertype)
	this.WriteHeadEnd(s)
}

type ResServerRegist struct {
	ProtocolInterface
	ProtocolHead
	Flag uint32
	Servertype uint32
}
func (this *ResServerRegist) Read(s *Biostream) {
	this.ReadHead(s)
	this.Flag = s.ReadUint32()
	this.Servertype = s.ReadUint32()
}
func (this *ResServerRegist) Write(s *Biostream) {
	this.WriteHead(s, XYID_RES_SERVERREGIST)
	s.WriteUint32(this.Flag)
	s.WriteUint32(this.Servertype)
	this.WriteHeadEnd(s)
}

type ReqLogin struct {
	ProtocolInterface
	ProtocolHead
	Account  string
	Password string
}
func (this *ReqLogin) Read(s *Biostream) {
	this.ReadHead(s)
	this.Account = s.ReadString()
	this.Password = s.ReadString()
}
func (this *ReqLogin) Write(s *Biostream) {
	this.WriteHead(s, XYID_REQ_LOGIN)
	s.WriteString(this.Account)
	s.WriteString(this.Password)
	this.WriteHeadEnd(s)
}

type ResLogin struct {
	ProtocolInterface
	ProtocolHead

	/*
		SUCCESS = 0,
		UNKNOWN,			// 未知错误
		NONE_NUMID,			// 数字id不存在
		PASSWORD_ERROR,		// 密码错误
		DB_ERROR,			// 数据库错误
		PS_STATE_ERROR,		// player状态错误
		PS_WAIT_CHECK,		// 未确认版本
		PS_WAIT_AUTH,		// 已经在登陆中了
		PS_ME_ONLINE,		// 我已经登陆过了(相同socket)
		PS_OTHER_ONLINE,	// 别人已经登录过了(不同socket)
	*/

	Flag     uint32
	Numidxy    uint32
	Nickname string
}
func (this *ResLogin) Read(s *Biostream) {
	this.ReadHead(s)
	this.Flag = s.ReadUint32()
	this.Numidxy = s.ReadUint32()
	this.Nickname = s.ReadString()
}
func (this *ResLogin) Write(s *Biostream) {
	this.WriteHead(s, XYID_RES_LOGIN)
	s.WriteUint32(this.Flag)
	s.WriteUint32(this.Numidxy)
	s.WriteString(this.Nickname)
	this.WriteHeadEnd(s)
}

type ReqPlayerLeave struct {
	ProtocolInterface
	ProtocolHead
}
func (this *ReqPlayerLeave) Read(s *Biostream) {
	this.ReadHead(s)
}
func (this *ReqPlayerLeave) Write(s *Biostream) {
	this.WriteHead(s, XYID_REQ_PLAYER_LEAVE)
	this.WriteHeadEnd(s)
}

type ReqRegist struct {
	ProtocolInterface
	ProtocolHead
	Nickname string
	Account  string
	Password string
}
func (this *ReqRegist) Read(s *Biostream) {
	this.ReadHead(s)
	this.Nickname = s.ReadString()
	this.Account = s.ReadString()
	this.Password = s.ReadString()
}
func (this *ReqRegist) Write(s *Biostream) {
	this.WriteHead(s, XYID_REQ_REGIST)
	s.WriteString(this.Nickname)
	s.WriteString(this.Account)
	s.WriteString(this.Password)
	this.WriteHeadEnd(s)
}

type ResRegist struct {
	ProtocolInterface
	ProtocolHead
	Flag uint32
	Numidxy uint32
}
func (this *ResRegist) Read(s *Biostream) {
	this.ReadHead(s)
	this.Flag = s.ReadUint32()
	this.Numidxy = s.ReadUint32()
}
func (this *ResRegist) Write(s *Biostream) {
	this.WriteHead(s, XYID_RES_REGIST)
	s.WriteUint32(this.Flag)
	s.WriteUint32(this.Numidxy)
	this.WriteHeadEnd(s)
}

type ReqJoinRoom struct {
	ProtocolInterface
	ProtocolHead
	GWid     uint32
	Nickname string
}
func (this *ReqJoinRoom) Read(s *Biostream) {
	this.ReadHead(s)
	this.GWid = s.ReadUint32()
	this.Nickname = s.ReadString()
}
func (this *ReqJoinRoom) Write(s *Biostream) {
	this.WriteHead(s, XYID_REQ_JOINROOM)
	s.WriteUint32(this.GWid)
	s.WriteString(this.Nickname)
	this.WriteHeadEnd(s)
}

type ResJoinRoom struct {
	ProtocolInterface
	ProtocolHead
	GWid     uint32
	Flag     uint32
	Nickname string
}
func (this *ResJoinRoom) Read(s *Biostream) {
	this.ReadHead(s)
	this.GWid = s.ReadUint32()
	this.Flag = s.ReadUint32()
	this.Nickname = s.ReadString()
}
func (this *ResJoinRoom) Write(s *Biostream) {
	this.WriteHead(s, XYID_RES_JOINROOM)
	s.WriteUint32(this.GWid)
	s.WriteUint32(this.Flag)
	s.WriteString(this.Nickname)
	this.WriteHeadEnd(s)
}

type ReqOpt struct {
	ProtocolInterface
	ProtocolHead
	FrameIndex uint32
	OptCnt uint32
	OptLen uint32
	Optdatas []byte
}
func (this *ReqOpt) Read(s *Biostream) {
	this.ReadHead(s)
	this.FrameIndex = s.ReadUint32()
	this.OptCnt = s.ReadUint32()
	this.OptLen = s.ReadUint32()
	this.Optdatas = s.ReadBytes(this.OptLen)
}
func (this *ReqOpt) Write(s *Biostream) {
	this.WriteHead(s, XYID_REQ_OPT)
	s.WriteUint32(this.FrameIndex)
	s.WriteUint32(this.OptCnt)
	s.WriteUint32(this.OptLen)
	s.WriteBytes(this.Optdatas, this.OptLen)
	this.WriteHeadEnd(s)
}

type NtfOpt struct {
	ProtocolInterface
	ProtocolHead
	Flag uint32
	FrameCnt uint32
	FrameLen uint32
	Framedatas []byte
}
func (this *NtfOpt) Read(s *Biostream) {
	this.ReadHead(s)
	this.Flag = s.ReadUint32()
	this.FrameCnt = s.ReadUint32()
	this.FrameLen = s.ReadUint32()
	this.Framedatas = s.ReadBytes(this.FrameLen)
}
func (this *NtfOpt) Write(s *Biostream) {
	this.WriteHead(s, XYID_NTF_OPT)
	s.WriteUint32(this.Flag)
	s.WriteUint32(this.FrameCnt)
	s.WriteUint32(this.FrameLen)
	s.WriteBytes(this.Framedatas, this.FrameLen)
	this.WriteHeadEnd(s)
}

type NtfGameStart struct {
	ProtocolInterface
	ProtocolHead
}
func (this *NtfGameStart) Read(s *Biostream) {
	this.ReadHead(s)
}
func (this *NtfGameStart) Write(s *Biostream) {
	this.WriteHead(s, XYID_NTF_GAMESTART)
	this.WriteHeadEnd(s)
}

type NtfMatching struct {
	ProtocolInterface
	ProtocolHead
	Flag uint32
}
func (this *NtfMatching) Read(s *Biostream) {
	this.ReadHead(s)
	this.Flag = s.ReadUint32()
}
func (this *NtfMatching) Write(s *Biostream) {
	this.WriteHead(s, XYID_NTF_MATCHING)
	s.WriteUint32(this.Flag)
	this.WriteHeadEnd(s)
}

type NtfGameEnd struct {
	ProtocolInterface
	ProtocolHead
}
func (this *NtfGameEnd) Read(s *Biostream) {
	this.ReadHead(s)
}
func (this *NtfGameEnd) Write(s *Biostream) {
	this.WriteHead(s, XYID_NTF_GAMEEND)
	this.WriteHeadEnd(s)
}

type ReqConnect struct {
	ProtocolInterface
	ProtocolHead
}
func (this *ReqConnect) Read(s *Biostream) {
	this.ReadHead(s)
}
func (this *ReqConnect) Write(s *Biostream) {
	this.WriteHead(s, XYID_REQ_CONNECT)
	this.WriteHeadEnd(s)
}