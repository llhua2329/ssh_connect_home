package main

import (
	"net"
	"fmt"
	"flag"
	"strconv"
	"time"
)

func main() {
	ip := flag.String("ip", "192.168.2.38", "remote ip")
	port := flag.Int("port", 50001, "remote port")
	flag.Parse()

	server_addr := *ip +":"+ strconv.Itoa(*port)
	control_conn, err := net.Dial("tcp", server_addr)
	if err != nil {
		return
	}

	defer control_conn.Close()

	for {
		byte := make([]byte, 1)
		len, err := control_conn.Read(byte)
		if err != nil {
			fmt.Println(time.Now(), err)
			return
		}

		if byte[0] == 0xFF { // 发起新的ssh连接
			fmt.Println(time.Now(), len, "accept now ssh")
			go newSsh(server_addr)
		}
	}
}

func newSsh(server_info string) {
	fmt.Println(time.Now(), "send server to create new ssh")
	remote, err := net.Dial("tcp", server_info)
	if err != nil {
		fmt.Println("err:", err)
		return
	}
	defer remote.Close()
	fmt.Println(time.Now(), "send local ssh")
	local, err := net.Dial("tcp", "127.0.0.1:22")
	if err != nil {
		fmt.Println("err:", err)
		return
	}
	fmt.Println(time.Now(), "swap data")
	defer local.Close()

	local.SetDeadline(time.Time{})
	remote.SetDeadline(time.Time{})
	go CopyConnection(local, remote)
	CopyConnection(remote, local)
}

func CopyConnection(dst net.Conn, src net.Conn) {
	defer dst.Close()
	defer src.Close()
	for {
		recvBuff := make([]byte, 1024)
		len, err := src.Read(recvBuff)
		if err != nil {
			fmt.Println("read info:", err)
			return
		}
		len, err = dst.Write(recvBuff[:len])
		if err != nil {
			fmt.Println("write info", err)
		}
	}
}