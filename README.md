
刚开始学习grpc，自己尝试写了简单的参考例子，可以用初学者学习。

- 将所有发送消息都通过bytes 进行发送，面得多方共同去维护proto文件，proto文件虽然相对于json压缩很多，但是维护版本着实烦躁。
- 可以同时作为服务器和客户端
- 支持同时连接多个服务器

## 包引入,当前版本 v1.0.0
```
go get github.com/Aiven37/MRpc@v1.0.0
```



## 创建服务器端：
```go
import (
    sever "github.com/Aiven37/MRpc/pkg/server"
)
server := sever.MRpcServer{}
handle := ServerHandle{}
err := server.StartServer(8787, &handle)
if err != nil {
    panic(err)
}
```
说明：
先示例话MRpcServer 对象，然后创建一个服务器消息接收器，这个接收器必须实现接口方法。然后传入端口启动服务。
需要实现的接口包括：
- Unary RPC （一元通信:最常见，像 HTTP 请求）:
```go
HandleRequest(sessionId string, msgId int32, data []byte) (int32, []byte, error)`
```
- Server Streaming RPC  (服务端流:客户端请求一次，服务端持续返回):
```go
 RequestMultResp(sessionId string, msgId int32, data []byte, ask *Asker) error`
```
- Client Streaming RPC(客户端连续发送，服务端一次返回):
```go
  MultiRequest(ch chan model.MsgReq) (int32, []byte, error)
```
- Bidirectional Streaming RPC (双向流,最强、也最复杂)
```go
  DuplexStream(ch chan model.MsgReq, ask *Asker) error
```
具体可参考examples 目录下的ServerMsgHandle 处理


## 创建客户端
*step1*： 创建客户端实例
 ```go
    client = mm_rpc_client.MRpcClient{}
``` 
*step2*： 添加服务器配置
```go
    //指定服务器IP及端口
defaultConfig := mm_rpc_model.ServerConfig{
    IP:   "127.0.0.1",
    Port: 8787,
}
// 添加连接配置，只是添加，并不会自动连接
client.AddConnect(&defaultConfig)

//client.AddConnect(&defaultConfig) 可以同时添加多个服务器连接
```
*step3*： 连接服务器，会将所有服务器配置的都启动
```go
  err = client.ConnectServer()
```

客户端提供接口有：
```go
    //一元通信
     Request(cfg model.ServerConfig, msgId int32, data []byte) (*model.MsgResp, error)
    //服务端流
     RequestMultResp(cfg model.ServerConfig, msgId int32, data []byte, callback model.MultRespCallback) error
    //客户端多请求，一次返回
     MultRequest(cfg model.ServerConfig) (*send.SReqeustSender, error)
    // 双向流
	 DuplexStream(cfg model.ServerConfig, callBack model.MultRespCallback) (*send.MReqeustSender, error)
```
客户端代码参照开源工程中的main.go中的代码


## License
This project is licensed under the MIT License.
