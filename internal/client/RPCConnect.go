package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	InModel "github.com/Aiven37/MRpc/internal/model"
	util "github.com/Aiven37/MRpc/internal/utils"
	pb "github.com/Aiven37/MRpc/message/v1"
	"github.com/Aiven37/MRpc/pkg/client/sender"
	model "github.com/Aiven37/MRpc/pkg/model"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RPCConnect struct {
	Config      *model.ServerConfig
	Client      pb.MRpcServiceClient
	Conn        *grpc.ClientConn
	IsConnected bool
}

func (conn *RPCConnect) StartConnection() error {
	util.LDebug("准备连接服务", zap.String("ip", conn.Config.IP), zap.Int("port", conn.Config.Port))
	cnn, err := grpc.NewClient(fmt.Sprintf("%s:%d", conn.Config.IP, conn.Config.Port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		util.LDebug("服务连接失败", zap.Error(err))
		return err
	}
	conn.Conn = cnn
	conn.Client = pb.NewMRpcServiceClient(cnn)
	conn.IsConnected = true
	util.LDebug("服务连接成功")
	return nil
}

func (conn *RPCConnect) StopConnection() error {
	if conn.IsConnected && conn.Conn != nil {
		_ = conn.Conn.Close()
		conn.Client = nil
	}
	return nil
}

func (conn *RPCConnect) Request(timeout time.Duration, msgId int32, data []byte) (*model.MsgResp, error) {
	if conn.IsConnected && conn.Conn != nil {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		in := &pb.MRpcMessage{
			MsgId:     msgId,
			Data:      data,
			SessionId: util.GenerateSessionId(32),
		}
		ack, err := conn.Client.Request(ctx, in)
		if err != nil {
			return nil, err
		}
		return &model.MsgResp{
			MsgId:     ack.AckMsgId,
			Data:      ack.AckData,
			SessionId: ack.SessionId,
		}, nil
	} else {
		return nil, errors.New("not connected")
	}
}

func (conn *RPCConnect) RequestMultResp(timeout time.Duration, msgId int32, data []byte, callback model.MultRespCallback) error {
	if conn.IsConnected && conn.Conn != nil {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		sid := util.GenerateSessionId(32)
		util.LDebug("发送RequestMultResp请求", zap.Int32("msgId", msgId), zap.String("sessionId", sid), zap.Int64("timeout", int64(timeout)))

		in := &pb.MRpcMessage{
			MsgId:     msgId,
			Data:      data,
			SessionId: sid,
		}
		stream, err := conn.Client.RequestMultResp(ctx, in)
		if err != nil {
			return err
		}
		for {
			ack, err := stream.Recv()
			if err == io.EOF {
				util.LDebug("RequestMultResp Recv 结束，到了EOF", zap.String("sessionId", sid))
				break
			}
			if err != nil {
				util.LError("RequestMultResp Recv 失败", zap.String("sessionId", sid), zap.Error(err))
				return err
			}
			callback(ack.SessionId, ack.AckMsgId, ack.AckData)
		}
	} else {
		return errors.New("not connected")
	}
	return nil
}

func (conn *RPCConnect) MultRequest(timeout time.Duration) (*sender.SReqeustSender, error) {
	if conn.IsConnected && conn.Conn != nil {
		sid := util.GenerateSessionId(32)
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		sd := sender.SReqeustSender{
			SessionId:  sid,
			CancelFunc: cancel,
		}
		util.LDebug("发送RequestMultResp请求", zap.String("sessionId", sid), zap.Int64("timeout", int64(timeout)))
		stream, err := conn.Client.MultRequest(ctx)
		if err != nil {
			util.LError("DuplexStream Recv 结束", zap.String("sessionId", sd.SessionId), zap.Int64("timeout", int64(timeout)))
			return nil, err
		}
		sd.Stream = stream
		return &sd, nil
	} else {
		return nil, errors.New("not connected")
	}
}

func (conn *RPCConnect) DuplexStream(callBack model.MultRespCallback, timeout time.Duration) (*sender.MReqeustSender, error) {
	if conn.IsConnected && conn.Conn != nil {
		//ctx, cancel := context.WithTimeout(context.Background(), timeout)
		ctx, cancel := context.WithCancel(context.Background())
		sd := sender.MReqeustSender{
			SessionId:  util.GenerateSessionId(32),
			CancelFunc: cancel,
		}
		util.LDebug("发送DuplexStream请求", zap.String("sessionId", sd.SessionId), zap.Int64("timeout", int64(timeout)))
		stream, err := conn.Client.DuplexStream(ctx)
		if err != nil {
			return nil, err
		}
		sd.Stream = stream
		go func(cbk model.MultRespCallback, send *sender.MReqeustSender) {
			for {
				in, err := stream.Recv()
				if err == io.EOF {
					send.Close()
					util.LDebug("DuplexStream Recv 结束，到了EOF", zap.String("sessionId", sd.SessionId))
					return
				}
				if err != nil {
					util.LError("DuplexStream Recv 失败", zap.String("sessionId", sd.SessionId), zap.Error(err))
					return
				}
				if in.GetAckMsgId() == InModel.MSG_SUE_EOF || in.GetAckMsgId() == InModel.MSG_SUE_ERR {
					send.Close()
					return
				}
				callBack(in.SessionId, in.GetAckMsgId(), in.GetAckData())
			}
		}(callBack, &sd)
		return &sd, nil
	} else {
		return nil, errors.New("not connected")
	}
}
