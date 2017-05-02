package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

type ProxyStruct struct {
	No                 int
	LastRecvTime       uint32
	ConnA              net.Conn
	ConnB              net.Conn
	LastSaveReportTime uint32
	LastRecvReportTime uint32
}

func proxyClientHandle(oneProxy *ProxyStruct) {
	defer oneProxy.ConnB.Close()
	buff := make([]byte, 1024*2)
	var buffUsed uint16 = 0
	for {
		length, err := oneProxy.ConnB.Read(buff[buffUsed:])
		if err != nil {
			fmt.Println("proxyClientHandle recv data err", err)
			return
		}
		fmt.Printf("Received:%d %v\r\n", length, buff[:length])
		_, err = oneProxy.ConnA.Write(buff[:length])
		if err != nil {
			//由于目标计算机积极拒绝而无法创建连接
			fmt.Println("转发失败", err.Error())
			return // 终止程序
		}
		buffUsed = 0


	}

}
func proxyConnectConnB(oneProxy *ProxyStruct){
	fmt.Println("proxyClientHandle start a connet to server!")
	server, err := net.Dial("tcp", "58.61.154.247:7018")
	if err != nil {
		//由于目标计算机积极拒绝而无法创建连接
		fmt.Println("Error dialing", err.Error())
		return // 终止程序
	}
	oneProxy.ConnB = server
	//defer connB.Close()
	fmt.Println("proxyClientHandle Connect Server Succese!")

	if err != nil {
		//由于目标计算机积极拒绝而无法创建连接
		fmt.Println("Error read", err.Error())
		return // 终止程序
	}

}

func proxyHandleConn(client net.Conn) {
	defer client.Close()
	oneProxy := ProxyStruct{}
	oneProxy.ConnA = client

	buff := make([]byte, 1024*4)

	fmt.Println("终端连接成功", client.RemoteAddr())
	proxyConnectConnB(&oneProxy)
	go proxyClientHandle(&oneProxy)

	for {
		client.SetReadDeadline(time.Now().Add(120 * time.Second))
		len, err := client.Read(buff[0:])
		if err != nil {
			log.Println("recv data err", err)
			return
		}
		log.Println("recv data, len:", len)
		if len > 0{
			oneProxy.ConnB.Write(buff)
			oneProxy.LastRecvTime = uint32(time.Now().Unix())
		}


		//log.Println("dump data:\n", hex.Dump(buff[buffUsed: buffUsed+uint16(len)]))

	}
}

func socketProxy() {

	listen_address := ":8095"
	logPath := "./detector_server.log"

	fmt.Println("server is starting...")
	logFile, logErr := os.OpenFile(logPath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if logErr != nil {
		log.Println("Fail to find", logPath, "cServer start Failed")
		os.Exit(1)
	}
	//logFile.Close()
	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	listen, err := net.Listen("tcp4", listen_address)
	if err != nil {
		return
	}
	fmt.Println("server start, listen on", listen_address)
	defer listen.Close()
	fmt.Println("server start done...")
	for {
		conn, err := listen.Accept()
		if err != nil {
			return
		}
		fmt.Println("accept new connection")
		go proxyHandleConn(conn);

	}
	//go proxyHandleConn(conn)
}

func main() {
	socketProxy()
}
