package main

import (
	"DrillServerGo/flog"
	"DrillServerGo/protocol"
	"DrillServerGo/socket"
	"fmt"
)

func (server *Server) gw_ReqLogin(io *nsocket.IOClientInterface, head protocol.ProtocolHead, buf []byte) {
	var req protocol.ReqLogin
	protocol.BufferToProtocol(buf, &req)
	flog.GetInstance().Infof("ReqLogin:account=%s", req.Account)

	var res protocol.ResLogin
	res.Numid = head.Numid

	sql := fmt.Sprintf("select numid,nickname,passwd from players where account='%s'", req.Account)
	row, err := server.mysql.Query(sql)
	if err != nil {
		flog.GetInstance().Errorf("ReqLogin:Select error:[%s],sql:[%s]", err, sql)
		res.Flag = 4
	} else {
		var numid uint32
		var nickname, passwd string
		if row.Next() { // 只有一行数据
			err = row.Scan(&numid, &nickname, &passwd)
			if err != nil {
				flog.GetInstance().Errorf("ReqLogin:Scan result error:%s", err)
				res.Flag = 4
			} else {
				if passwd == req.Password {
					res.Flag = 0
					res.Numidxy = numid
					res.Nickname = nickname
					flog.GetInstance().Debugf("ReqLogin:Login success")
				} else {
					flog.GetInstance().Debugf("ReqLogin:Password error,account=%s", req.Account)
					res.Flag = 3
				}
			}
		} else {
			flog.GetInstance().Warnf("ReqLogin:No data,account=%s", req.Account)
			res.Flag = 1
		}
	}

	sendbuf := make([]byte, 128)
	len := protocol.ProtocolToBuffer(&res, sendbuf)
	(*io).Send(sendbuf[:len])
}

func (server *Server) gw_ReqRegist(io *nsocket.IOClientInterface, head protocol.ProtocolHead, buf []byte) {
	var req protocol.ReqRegist
	protocol.BufferToProtocol(buf, &req)
	flog.GetInstance().Infof("ReqRegist:account=%d,nickname=%s", req.Account, req.Nickname)

	var res protocol.ResRegist
	res.Numid = req.Numid // 这里用的是tempid

	sql := fmt.Sprintf("insert into players (account,nickname,passwd) values ('%s','%s','%s')", req.Account, req.Nickname, req.Password)
	result, err := server.mysql.Exec(sql)
	if err != nil {
		flog.GetInstance().Errorf("ReqRegist:insert error:%s", err)
		res.Flag = 7
	} else {
		cnt, _ := result.RowsAffected()
		if cnt != 1 {
			flog.GetInstance().Infof("ReqRegist:insert affected=%d,account=%s,nickname=%s", cnt, req.Account, req.Nickname)
			res.Flag = 1
		} else {
			// 再查询生成的numid
			sql = fmt.Sprintf("select numid from players where account='%s'", req.Account)
			row, err := server.mysql.Query(sql)
			if err != nil {
				flog.GetInstance().Errorf("ReqRegist:Select numid error.account=%s,error:%s", req.Account, err)
				res.Flag = 6
			} else {
				if row.Next() {
					var numid uint32
					err = row.Scan(numid)
					if err != nil {
						flog.GetInstance().Errorf("ReqRegist:Scan error:%s", err)
						res.Flag = 6
					} else {
						res.Numidxy = numid
						res.Flag = 0
					}
				} else {
					flog.GetInstance().Errorf("ReqRegist:Select numid no data,account=%s", req.Account)
					res.Flag = 6
				}
			}
		}
	}

	sendbuf := make([]byte, 128)
	len := protocol.ProtocolToBuffer(&res, sendbuf)
	(*io).Send(sendbuf[:len])
}

func (server *Server) ReciveFromGW(io *nsocket.IOClientInterface, head protocol.ProtocolHead, buf []byte) {
	switch head.Xyid {
	case protocol.XYID_REQ_LOGIN:
		server.gw_ReqLogin(io, head, buf)
	case protocol.XYID_REQ_REGIST: // 未测
		server.gw_ReqRegist(io, head, buf)
	default:
		flog.GetInstance().Error("Unknown xyid=%d,len=%d", head.Xyid, head.Length)
		break
	}
}
