package mm_rpc_model

import "time"

type ServerConfig struct {
	IP                  string
	Port                int
	SRequestTimeOut     time.Duration
	MRequestTimeOut     time.Duration
	MRStreamTimeOut     time.Duration
	MMultiStreamTimeOut time.Duration
}

type MsgReq struct {
	SessionId string `json:"sessionId"`
	MsgId     int32  `json:"msgId"`
	Data      []byte `json:"data"`
}

type MsgResp struct {
	SessionId string `json:"sessionId"`
	MsgId     int32  `json:"msgId"`
	Data      []byte `json:"data"`
}

type MultRespCallback func(string, int32, []byte)
