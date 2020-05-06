package main

// 暂时废弃了,无法编译

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"DrillServerGo/flog"
	"DrillServerGo/protocol"
)

var ClientCount = flag.Int("c",10,"How much clients")
var LoopTimes = flag.Int("l",100,"Loop times per socket")
var WaitTime = flag.Int("w",0,"Wait time(msec) per Heartbeat")
var addr1 = flag.String("addr", "127.0.0.1:6001", "http service address")

// top -H -d 0.5 -p `pgrep "Web|server" |xargs perl -e "print join ',',@ARGV"`

func runOneClient(looptimes int, wg *sync.WaitGroup) {
	u := url.URL{Scheme:"ws", Host:*addr1, Path:""}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		flog.GetInstance().Errorf("connect error,err=", err)
		wg.Done()
		return
	}

	defer func() {
		c.Close()
		wg.Done()
	}()

	for nowloop := 0; looptimes == -1 || nowloop < looptimes;  nowloop++ {
		// 发送心跳
		var resp protocol.HeartBeat
		resp.Timestamp = uint32(time.Now().Unix())
		flog.GetInstance().Debugf("send:timestamp=%d", resp.Timestamp)
		sendbuf := resp.Encode()
		err := c.WriteMessage(websocket.BinaryMessage, sendbuf)
		if err != nil {
			flog.GetInstance().Error("write err:", err)
			return
		}

		// 收到心跳
		_, recvbuf, err := c.ReadMessage()
		if err != nil {
			flog.GetInstance().Error("read err:", err)
			return
		}
		var req protocol.HeartBeat
		if(req.GetHeadAndAttach(recvbuf)) {
			req.Decode()
			flog.GetInstance().Debugf("recv:timestamp=%d", req.Timestamp)
		} else {
			flog.GetInstance().Errorf("recv a error message")
			return
		}

		if *WaitTime >0 {
			time.Sleep(time.Millisecond * time.Duration(*WaitTime))
		}
	}

	_ = c.WriteMessage(websocket.CloseMessage, []byte{})
}

func main() {
	err := flog.InitLog("client", flog.LevelDebug)
	if err != nil {
		fmt.Println(os.Stderr, "InitLog error:", err)
		os.Exit(1)
	}

	flag.Parse()
	flog.GetInstance().Infof("ClientCount=%d,LoopTimes=%d,WaitTime=%d,addr=%s",
						*ClientCount, *LoopTimes, *WaitTime, *addr1)

	begin := time.Now().UnixNano()  / 1e6
	var wg sync.WaitGroup
	for i:=0; i<*ClientCount ; i++ {
		wg.Add(1)
		go runOneClient(*LoopTimes, &wg)
	}
	wg.Wait()
	end := time.Now().UnixNano()  / 1e6
	flog.GetInstance().Debug("usetimes=", (end-begin))
}
