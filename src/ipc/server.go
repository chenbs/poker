/*
IPC简单框架的服务器端
IPC框架用来封装通信包的编码细节，用channel作为模块之间的通信方式，
传递数据用Json格式的字符串类型
*/

package ipc

import (
	"encoding/json"
	"fmt"
)

type Request struct {
	Method string "method"
	Params string "params"
	//Room   int    "room"
}

type Response struct {
	Code string "code"
	Body string "body"
}

type Server interface {
	Name() string
	Handle(method, params string) *Response
}

type IpcServer struct {
	Server
}

func NewIpcServer(server Server) *IpcServer {
	return &IpcServer{server}
}

func (server *IpcServer) Connect() chan string {
	session := make(chan string, 0)

	go func(c chan string) {
		for {
			request := <-c

			if request == "CLOSE" { //关闭该连接
				break
			}

			var req Request
			err := json.Unmarshal([]byte(request), &req)
			if err != nil {
				fmt.Println("Invalid reuqset format:", request)
			}

			resp := server.Handle(req.Method, req.Params)

			b, err := json.Marshal(resp)

			c <- string(b) //返回结果
		}

		fmt.Println("Session closed")
	}(session)

	fmt.Println("A new session has been created successfully.")

	return session
}
