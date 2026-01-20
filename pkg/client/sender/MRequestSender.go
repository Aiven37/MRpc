package sender

import (
	"context"

	util "github.com/Aiven37/MRpc/internal/utils"
	pb "github.com/Aiven37/MRpc/message/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type MReqeustSender struct {
	SessionId  string
	CancelFunc context.CancelFunc
	Stream     grpc.BidiStreamingClient[pb.MRpcMessage, pb.MRpcMessageAck]
}

func (s *MReqeustSender) SendMessage(messageId int32, msg []byte) error {
	util.LDebug("发送消息", zap.String("sessionId", s.SessionId))
	return s.Stream.Send(&pb.MRpcMessage{
		SessionId: s.SessionId,
		MsgId:     messageId,
		Data:      msg,
	})
}

func (s *MReqeustSender) SendOver() error {
	util.LDebug("CloseSend", zap.String("sessionId", s.SessionId))
	err := s.Stream.CloseSend()
	if err != nil {
		s.CancelFunc()
		s.CancelFunc = nil
		return err
	}
	return nil
}

func (s *MReqeustSender) Close() {
	util.LDebug("Close关闭")
	if s.CancelFunc != nil {
		s.CancelFunc()
	}
	s.CancelFunc = nil
}
