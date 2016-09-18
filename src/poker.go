package main

import (
	"fmt"
	"net"
)

// 用来记录所有的客户端连接
var ConnMap map[string]*net.TCPConn

//主函数，用来开启服务器，接收连接
func main() {

	fmt.Println("Starting the server ...")
	fmt.Println("-----version:1.0------")
	StartCenterService()
	var tcpAddr *net.TCPAddr
	ConnMap = make(map[string]*net.TCPConn)
	tcpAddr, _ = net.ResolveTCPAddr("tcp", "192.168.100.58:9999")
	tcpListener, err := net.ListenTCP("tcp", tcpAddr)
	defer tcpListener.Close()

	if err != nil {
		fmt.Println("Server starting failed...")
		panic(err)
	} else {
		fmt.Println("server is OK ,ready to connect")
	}

	for {
		tcpConn, err := tcpListener.AcceptTCP()
		if err != nil {
			continue
		}
		fmt.Println("A client connected... : " + tcpConn.RemoteAddr().String())
		go TcpPipe(tcpConn)
	}
}
