package main

import (
	"net"
	"fmt"
	"flag"
	"strconv"
)

func Copy(dst net.Conn, src net.Conn) {
	defer dst.Close()
	defer src.Close()
	for {
		recvBuff := make([]byte, 1024)
		len, err := src.Read(recvBuff)
		if err != nil {
			fmt.Println("Copy err:", err)
			return
		}
		dst.Write(recvBuff[:len])
	}
}

func main() {
	ip := flag.String("ip", "192.168.2.38", "remote ip")
	port := flag.Int("port", 12346, "remote port")
	flag.Parse()

	server_addr := *ip +":"+ strconv.Itoa(*port)
	control_conn, err := net.Dial("tcp", server_addr)
	if err != nil {
		return
	}
	defer control_conn.Close()

	for {
		byte := make([]byte, 1)
		_, err := control_conn.Read(byte)
		if err != nil {
			return
		}

		if (byte[0] == 0xFF) { // 发起新的ssh连接
			go newSsh(server_addr)
		}
	}
}

func newSsh(server_info string) {
	remote, err := net.Dial("tcp", server_info)
	if err != nil {
		fmt.Println("err:", err)
		return
	}
	defer remote.Close()
	local, err := net.Dial("tcp", "127.0.0.1:22")
	if err != nil {
		fmt.Println("err:", err)
		return
	}
	defer local.Close()

	go Copy(local, remote)
	Copy(remote, local)
}