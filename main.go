package main

import (
	"sync"
	"time"

	wtest "github.com/Aiven37/MRpc/examples"
	util "github.com/Aiven37/MRpc/internal/utils"
	mm_rpc_client "github.com/Aiven37/MRpc/pkg/client"
	mm_rpc_model "github.com/Aiven37/MRpc/pkg/model"
	mm_rpc_sever "github.com/Aiven37/MRpc/pkg/server"
	"go.uber.org/zap"
)

var client mm_rpc_client.MRpcClient

func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		doTest()
	}()
	wg.Add(1)
	wg.Wait()
}

func doTest() {
	server := mm_rpc_sever.MRpcServer{}
	shandle := wtest.ServerHandle{}
	server.OpenLogOut("output/log/")
	err := server.StartServer(8787, &shandle)
	if err != nil {
		util.LError(err.Error())
		return
	}

	defaultConfig := mm_rpc_model.ServerConfig{
		IP:   "127.0.0.1",
		Port: 8787,
	}

	time.Sleep(1 * time.Second)
	client = mm_rpc_client.MRpcClient{}
	client.OpenLogOut("output/log/")

	err = client.AddConnect(&defaultConfig)
	if err != nil {
		util.LError(err.Error())
		return
	}

	err = client.ConnectServer()
	if err != nil {
		util.LError(err.Error())
		return
	}

	testSingleRequest(defaultConfig)

	testMultiReqSingleResp(defaultConfig)

	testSingleReqMultiResp(defaultConfig)

	testDuplexStream(defaultConfig)

}

func testSingleRequest(config mm_rpc_model.ServerConfig) {
	util.LDebug("==============================================================================================================================================")
	//模拟客户端向服务器发送请求
	resp, err := client.Request(config, 1000, []byte("我是客户端"))
	if err != nil {
		util.LError(err.Error())
	} else {
		util.LDebug("收到服务器返回", zap.String("sessionId", resp.SessionId), zap.Int32("msgId", resp.MsgId), zap.String("data", string(resp.Data)))
	}
}

func testMultiReqSingleResp(config mm_rpc_model.ServerConfig) {
	util.LDebug("==============================================================================================================================================")
	sd, err := client.MultRequest(config)
	if err != nil {
		util.LError(err.Error())
	} else {
		err = sd.SendMessage(20001, []byte("我是接口多请求数据1"))
		if err != nil {
			util.LError(err.Error())
		}
		err = sd.SendMessage(20002, []byte("我是接口多请求数据2"))
		if err != nil {
			util.LError(err.Error())
		}
		err = sd.SendMessage(20003, []byte("我是接口多请求数据3"))
		if err != nil {
			util.LError(err.Error())
		}
		ack, err := sd.SendOver()
		if err != nil {
			util.LError(err.Error())
		} else {
			util.LDebug("MultRequest 收到服务器返回", zap.String("sessionId", ack.SessionId), zap.Int32("msgId", ack.AckMsgId), zap.String("data", string(ack.AckData)))
		}
	}
}

func testSingleReqMultiResp(config mm_rpc_model.ServerConfig) {
	util.LDebug("==============================================================================================================================================")
	err := client.RequestMultResp(config, 30001, []byte("我是一次请求多次回调"), func(s string, i int32, bytes []byte) {
		util.LDebug("RequestMultResp 收到服务器返回", zap.String("sessionId", s), zap.Int32("msgId", i), zap.String("data", string(bytes)))
	})
	if err != nil {
		util.LError(err.Error())
	}
}

// 测试双向流
func testDuplexStream(config mm_rpc_model.ServerConfig) {
	util.LDebug("==============================================================================================================================================")
	send, err := client.DuplexStream(config, func(sessionId string, msgId int32, data []byte) {
		util.LDebug("DuplexStream 收到服务器返回", zap.String("sessionId", sessionId), zap.Int32("msgId", msgId), zap.String("data", string(data)))
	})
	if err != nil {
		util.LError(err.Error())
	} else {
		err = send.SendMessage(40001, []byte("这是客户端发送给 DuplexStream 方法的数据 1 "))
		if err != nil {
			util.LError(err.Error())
		}
		time.Sleep(1 * time.Second)
		err = send.SendMessage(40002, []byte("这是客户端发送给 DuplexStream 方法的数据 2 "))
		if err != nil {
			util.LError(err.Error())
		}

		time.Sleep(1 * time.Second)
		err = send.SendMessage(40003, []byte("这是客户端发送给 DuplexStream 方法的数据 3 "))
		if err != nil {
			util.LError(err.Error())
		}
		time.Sleep(1 * time.Second)
		err = send.SendOver()
		if err != nil {
			util.LError(err.Error())
		}
	}
}
