package wtest

import (
	"fmt"
	"time"

	util "github.com/Aiven37/MRpc/internal/utils"
	model "github.com/Aiven37/MRpc/pkg/model"
	mm_rpc_sever "github.com/Aiven37/MRpc/pkg/server"
	"go.uber.org/zap"
)

type ServerHandle struct {
}

func (s *ServerHandle) HandleRequest(sessionId string, msgId int32, data []byte) (int32, []byte, error) {
	util.LDebug("Server: 收到客户端请求", zap.String("sessionId", sessionId), zap.Int32("msgId", msgId), zap.String("data", string(data)))
	return msgId, []byte("我是服务器端返回的数据"), nil
}

func (s *ServerHandle) RequestMultResp(sessionId string, msgId int32, data []byte, ask *mm_rpc_sever.Asker) error {
	util.LDebug("Server: 收到客户端请求", zap.String("sessionId", sessionId), zap.Int32("msgId", msgId), zap.String("data", string(data)))
	for i := int32(0); i < 5; i++ {
		err := ask.Ask(msgId+1+i, []byte(fmt.Sprintf("我是 RequestMultResp 消息ID：%d", msgId+1+i)))
		time.Sleep(1 * time.Second)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *ServerHandle) MultiRequest(ch chan model.MsgReq) (int32, []byte, error) {
	for c := range ch {
		util.LDebug("Server: 收到客户端请求", zap.String("sessionId", c.SessionId), zap.Int32("msgId", c.MsgId), zap.String("data", string(c.Data)))
	}
	return -10000, []byte("我是服务器端 MultiRequest 返回数据"), nil
}

func (s *ServerHandle) DuplexStream(ch chan model.MsgReq, ask *mm_rpc_sever.Asker) error {
	for c := range ch {
		util.LDebug("Server: 收到客户端请求", zap.String("sessionId", c.SessionId), zap.Int32("msgId", c.MsgId), zap.String("data", string(c.Data)))
		err := ask.Ask(c.MsgId, []byte("我是服务器端回复数据"))
		if err != nil {
			return err
		}
	}
	return nil
}
