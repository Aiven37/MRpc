package sender

import (
	"context"

	util "github.com/Aiven37/MRpc/internal/utils"
	pb "github.com/Aiven37/MRpc/message/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type SReqeustSender struct {
	SessionId  string
	CancelFunc context.CancelFunc
	Stream     grpc.ClientStreamingClient[pb.MRpcMessage, pb.MRpcMessageAck]
}

func (s *SReqeustSender) SendMessage(messageId int32, msg []byte) error {
	util.LDebug("发送消息", zap.String("sessionId", s.SessionId))
	return s.Stream.Send(&pb.MRpcMessage{
		SessionId: s.SessionId,
		MsgId:     messageId,
		Data:      msg,
	})
}

func (s *SReqeustSender) SendOver() (*pb.MRpcMessageAck, error) {
	defer s.CancelFunc()
	ack, err := s.Stream.CloseAndRecv()
	if err != nil {
		return nil, err
	}
	return ack, nil
}
