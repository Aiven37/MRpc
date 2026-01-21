package mm_rpc_client

import (
	"errors"
	"fmt"
	"time"

	"github.com/Aiven37/MRpc/internal/client"
	util "github.com/Aiven37/MRpc/internal/utils"
	send "github.com/Aiven37/MRpc/pkg/client/sender"
	model "github.com/Aiven37/MRpc/pkg/model"
	"go.uber.org/zap"
)

type MRpcClient struct {
	connects map[string]*client.RPCConnect
}

func (h *MRpcClient) fillDefaultConfig(config *model.ServerConfig) {
	if config.SRequestTimeOut == 0 {
		config.SRequestTimeOut = time.Second * 5
	}
	if config.MRequestTimeOut == 0 {
		config.MRequestTimeOut = time.Second * 120
	}
	if config.MRStreamTimeOut == 0 {
		config.MRStreamTimeOut = time.Second * 120
	}
	if config.MMultiStreamTimeOut == 0 {
		config.MMultiStreamTimeOut = time.Second * 600
	}
}

func (h *MRpcClient) buildKey(config *model.ServerConfig) string {
	return fmt.Sprintf("%s:%d", config.IP, config.Port)
}

func (h *MRpcClient) AddConnect(config *model.ServerConfig) error {
	if h.connects == nil {
		h.connects = make(map[string]*client.RPCConnect)
	}
	h.fillDefaultConfig(config)
	_, exist := h.connects[h.buildKey(config)]
	if exist {
		return errors.New("server " + h.buildKey(config) + " already exist")
	}
	util.LDebug("添加连接服务", zap.String("ip", config.IP), zap.Int("port", config.Port))
	h.connects[h.buildKey(config)] = &client.RPCConnect{
		Config: config,
	}
	return nil
}

func (h *MRpcClient) AddServerAndConnect(config *model.ServerConfig) error {
	_ = h.AddConnect(config)
	conn, exist := h.connects[h.buildKey(config)]
	if exist {
		if conn.IsConnected {
			return errors.New("connect " + h.buildKey(config) + " already connection")
		} else {
			return conn.StartConnection()
		}
	} else {
		return errors.New("server " + h.buildKey(config) + " not exist")
	}
}

func (h *MRpcClient) ConnectServer() error {
	for _, conn := range h.connects {
		if err := conn.StartConnection(); err != nil {
			return err
		}
	}
	return nil
}

func (h *MRpcClient) StopConnectServer() error {
	var globalError error = nil
	for _, server := range h.connects {
		err := server.StopConnection()
		if err != nil {
			if globalError == nil {
				globalError = err
			} else {
				globalError = fmt.Errorf("%s - %s", globalError.Error(), err.Error())
			}
		}
	}
	return globalError
}

func (h *MRpcClient) findServerConnect(cfg model.ServerConfig) (*client.RPCConnect, error) {
	for _, conn := range h.connects {
		if h.buildKey(conn.Config) == h.buildKey(&cfg) {
			return conn, nil
		}
	}
	return nil, errors.New("not found server config")
}

func (h *MRpcClient) Request(cfg model.ServerConfig, msgId int32, data []byte) (*model.MsgResp, error) {
	conn, err := h.findServerConnect(cfg)
	if err != nil {
		return nil, err
	} else {
		return conn.Request(conn.Config.SRequestTimeOut, msgId, data)
	}
}

func (h *MRpcClient) RequestMultResp(cfg model.ServerConfig, msgId int32, data []byte, callback model.MultRespCallback) error {
	conn, err := h.findServerConnect(cfg)
	if err != nil {
		return err
	} else {
		return conn.RequestMultResp(conn.Config.MRStreamTimeOut, msgId, data, callback)
	}
}

func (h *MRpcClient) MultRequest(cfg model.ServerConfig) (*send.SReqeustSender, error) {
	conn, err := h.findServerConnect(cfg)
	if err != nil {
		return nil, err
	} else {
		return conn.MultRequest(conn.Config.MRequestTimeOut)
	}
}

func (h *MRpcClient) DuplexStream(cfg model.ServerConfig, callBack model.MultRespCallback) (*send.MReqeustSender, error) {
	conn, err := h.findServerConnect(cfg)
	if err != nil {
		return nil, err
	} else {
		return conn.DuplexStream(callBack, conn.Config.MMultiStreamTimeOut)
	}
}

func (h *MRpcClient) OpenLogOut(dir string) {
	util.InitLogger(dir)
}
