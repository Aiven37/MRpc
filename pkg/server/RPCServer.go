package mm_rpc_sever

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"

	inMode "github.com/Aiven37/MRpc/internal/model"
	util "github.com/Aiven37/MRpc/internal/utils"
	pb "github.com/Aiven37/MRpc/message/v1"
	model "github.com/Aiven37/MRpc/pkg/model"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type MRpcServer struct {
	pb.UnimplementedMRpcServiceServer
	msgHandler *IServerMsgHandle
	grpcServer *grpc.Server
	port       int
	isRunning  bool
	Context    context.Context
	CancelFunc context.CancelFunc
}

func (s *MRpcServer) StartServer(port int, handle IServerMsgHandle) error {
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		return err
	}
	s.port = port
	s.Context, s.CancelFunc = context.WithCancel(context.Background())
	go func() {
		defer s.CancelFunc()
		util.LDebug("prepare start server", zap.String("port", fmt.Sprintf("%d", s.port)))
		s.grpcServer = grpc.NewServer()
		s.msgHandler = &handle
		pb.RegisterMRpcServiceServer(s.grpcServer, s)
		s.isRunning = true
		err = s.grpcServer.Serve(lis)
		if err != nil {
			util.LDebug("服务监听中断", zap.String("port", fmt.Sprintf("%d", s.port)))
		}
	}()
	return err
}

func (s *MRpcServer) StopServer() error {
	util.LDebug("server:rpc stop server", zap.String("port", fmt.Sprintf("%d", s.port)))
	defer func() {
		s.isRunning = false
	}()
	if s.isRunning && s.grpcServer != nil {
		s.grpcServer.Stop()
	}
	return nil
}

func (s *MRpcServer) Request(_ context.Context, msg *pb.MRpcMessage) (*pb.MRpcMessageAck, error) {
	util.LDebug("server:rpc server received request", zap.String("msg", msg.SessionId))
	if s.msgHandler != nil {
		msgId, data, err := (*s.msgHandler).HandleRequest(msg.SessionId, msg.MsgId, msg.Data)
		if err != nil {
			return nil, err
		}
		return &pb.MRpcMessageAck{
			SessionId: msg.SessionId,
			AckMsgId:  msgId,
			AckData:   data,
		}, nil
	} else {
		return nil, errors.New("msgHandler is nil")
	}
}

func (s *MRpcServer) RequestMultResp(msg *pb.MRpcMessage, stream grpc.ServerStreamingServer[pb.MRpcMessageAck]) error {
	util.LDebug("server: server: rpc server received RequestMultResp", zap.String("msg", msg.SessionId))
	if s.msgHandler != nil {
		asker := &Asker{
			IsMult:        false,
			MAskServer:    stream,
			MStreamServer: nil,
			SessionId:     msg.SessionId,
		}
		err := (*s.msgHandler).RequestMultResp(msg.SessionId, msg.MsgId, msg.Data, asker)
		return err
	} else {
		return errors.New("msgHandler is nil")
	}
}

func (s *MRpcServer) MultRequest(stream grpc.ClientStreamingServer[pb.MRpcMessage, pb.MRpcMessageAck]) error {
	util.LDebug("server: rpc server received MultRequest")
	var ch = make(chan model.MsgReq)
	go func() {
		defer close(ch)
		for {
			msg, err := stream.Recv()
			if err == io.EOF {
				return
			} else if err != nil {
				return
			} else {
				req := model.MsgReq{
					SessionId: msg.SessionId,
					MsgId:     msg.MsgId,
					Data:      msg.Data,
				}
				ch <- req
			}
		}
	}()
	var globalErr error = nil
	var resp *pb.MRpcMessageAck
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		respMsgId, respData, err := (*s.msgHandler).MultiRequest(ch)
		if err != nil {
			globalErr = err
		} else {
			resp = &pb.MRpcMessageAck{
				SessionId: "",
				AckMsgId:  respMsgId,
				AckData:   respData,
			}
		}
	}()
	wg.Wait()
	if globalErr != nil {
		return globalErr
	} else {
		return stream.SendAndClose(resp)
	}
}

func (s *MRpcServer) DuplexStream(stream grpc.BidiStreamingServer[pb.MRpcMessage, pb.MRpcMessageAck]) error {
	util.LDebug("server: rpc server received DuplexStream")
	if s.msgHandler != nil {
		asker := &Asker{
			IsMult:        false,
			MAskServer:    nil,
			MStreamServer: stream,
			SessionId:     "",
			Ch:            make(chan model.MsgReq),
		}
		go func(ack *Asker) {
			defer close(ack.Ch)
			for {
				msg, err := stream.Recv()
				if err == io.EOF {
					util.LDebug("server: rpc server received EOF ,发送确认信息给客户端关闭连接")
					_ = ack.Ask(inMode.MSG_SUE_EOF, []byte("--EOF--"))
					return
				} else if err != nil {
					util.LDebug("server: rpc server received ERROR ,发送确认信息给客户端关闭连接")
					_ = ack.Ask(inMode.MSG_SUE_ERR, []byte("--ERROR--"))
					return
				} else {
					asker.SessionId = msg.SessionId
					req := model.MsgReq{
						SessionId: msg.SessionId,
						MsgId:     msg.MsgId,
						Data:      msg.Data,
					}
					ack.Ch <- req
				}
			}
		}(asker)
		var globalErr error = nil
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := (*s.msgHandler).DuplexStream(asker.Ch, asker)
			if err != nil {
				globalErr = err
			}
		}()
		wg.Wait()
		return globalErr
	} else {
		return errors.New("msgHandler is nil")
	}
}

func (s *MRpcServer) OpenLogOut(dir string) {
	util.InitLogger(dir)
}
