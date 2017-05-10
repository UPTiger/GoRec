package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
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
func proxyConnectConnB(oneProxy *ProxyStruct) {
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
		if len > 0 {
			oneProxy.ConnB.Write(buff)
			oneProxy.LastRecvTime = uint32(time.Now().Unix())
		}

		//log.Println("dump data:\n", hex.Dump(buff[buffUsed: buffUsed+uint16(len)]))

	}
}

func socketProxy() {
	return
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
		go proxyHandleConn(conn)

	}
	//go proxyHandleConn(conn)
}

//----------------------------------------------
//自定义数据模式，Header+Frame
type LogReader struct {
	conn   net.Conn
	reader *bufio.Reader
	buffer [4]byte
}

func NewLogReader(c net.Conn) *LogReader {
	return &LogReader{
		conn:   c,
		reader: bufio.NewReader(c),
	}
}
func (p *LogReader) readHeader() (int32, error) {
	buf := p.buffer[:4]
	fmt.Println("==readHeader==")

	if _, err := io.ReadFull(p.reader, buf); err != nil {
		fmt.Println("readHeader error")
		return 0, err
	}
	size := int32(binary.BigEndian.Uint32(buf))
	if size < 0 || size > 1638400 {
		return 0, fmt.Errorf("Incorrect frame size(%d)", size)
	}
	fmt.Println("readHeader size:%d", size)
	return size, nil
}
func (p *LogReader) readFrame(size int) ([]byte, error) {
	var buf []byte
	if size <= len(p.buffer) {
		buf = p.buffer[0:size]
	} else {
		buf = make([]byte, size)
	}
	_, err := io.ReadFull(p.reader, buf)
	fmt.Println(buf)
	return buf, err
}

//数据写入队列
func forwardMessage(c net.Conn, queue chan<- string) {
	defer c.Close()
	logReader := NewLogReader(c)
	for {
		size, err := logReader.readHeader()
		if err != nil {
			break
		}
		data, err := logReader.readFrame(int(size))
		if err != nil {
			break
		}
		queue <- string(data)
	}
}

//从队列中读取数据、定期处理
func processMsg(q <-chan string) {
	ticker := time.NewTicker(time.Second * 60)

	for {
		select {
		case msg := <-q:
			fmt.Println(msg)
		case <-ticker.C:
			fmt.Println("定时任务")

		}
	}
}
func serverTest() {
	address := ":8095"
	listen, err := net.Listen("tcp4", address)
	if err != nil {
		return
	}
	fmt.Println("server start, listen on", address)
	defer listen.Close()
	fmt.Println("server start done...")

	messageQueue := make(chan string, 40960)
	go processMsg(messageQueue)
	for {
		fd, err := listen.Accept()
		if err != nil {

			continue
		}
		fmt.Println("accept new connection")
		go forwardMessage(fd, messageQueue)

	}

}

//----------------------------------------------
func main() {
	serverTest()

	socketProxy()
}
