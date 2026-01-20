package mm_rpc_sever

import (
	"errors"

	pb "github.com/Aiven37/MRpc/message/v1"
	model "github.com/Aiven37/MRpc/pkg/model"
	"google.golang.org/grpc"
)

type Asker struct {
	IsMult        bool
	SessionId     string
	Ch            chan model.MsgReq
	MAskServer    grpc.ServerStreamingServer[pb.MRpcMessageAck]
	MStreamServer grpc.BidiStreamingServer[pb.MRpcMessage, pb.MRpcMessageAck]
}

func (asker *Asker) Ask(msgId int32, data []byte) error {
	if !asker.IsMult && asker.MAskServer != nil {
		return asker.MAskServer.Send(&pb.MRpcMessageAck{
			SessionId: asker.SessionId,
			AckMsgId:  msgId,
			AckData:   data,
		})
	} else if asker.MStreamServer != nil {
		return asker.MStreamServer.Send(&pb.MRpcMessageAck{
			SessionId: asker.SessionId,
			AckMsgId:  msgId,
			AckData:   data,
		})
	}
	return errors.New("system error")
}
