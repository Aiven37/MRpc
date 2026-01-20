package mm_rpc_sever

import (
	model "github.com/Aiven37/MRpc/pkg/model"
)

type IServerMsgHandle interface {
	/**
	*  客户端相应请求
	 */
	HandleRequest(sessionId string, msgId int32, data []byte) (int32, []byte, error)

	/**
	* 客户端单请求，多次Resp
	 */
	RequestMultResp(sessionId string, msgId int32, data []byte, ask *Asker) error

	/**
	 * 客户端多请求，单次Resp
	 */
	MultiRequest(ch chan model.MsgReq) (int32, []byte, error)

	/**
	* 双向流
	 */
	DuplexStream(ch chan model.MsgReq, ask *Asker) error
}
