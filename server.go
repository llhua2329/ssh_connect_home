package main

import (
	"net"
	"fmt"
	"flag"
	"strconv"
)

func main() {
	user_port := flag.Int("user_port", 12345, "user connet port")
	home_port := flag.Int("home_port", 12346, "home connect_port")
	flag.Parse()
	listen_user, err := net.Listen("tcp", "0.0.0.0:"+ strconv.Itoa(*user_port))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer listen_user.Close()

	listen_home, err := net.Listen("tcp", "0.0.0.0:" + strconv.Itoa(*home_port))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer listen_home.Close()

	control_conn, err := listen_home.Accept()
	if err != nil {
		fmt.Println("control connect failed", err)
		return
	}
	defer control_conn.Close()

	for {
		user_conn, err := listen_user.Accept()
		if err !=nil {
			continue
		}

		byte := make([]byte, 1)
		byte[0] = 0xFF  // 通知home 有新的ssh来连接
		control_conn.Write(byte)

		home_conn, err := listen_home.Accept() // 等待home来连接
		if err != nil {
			user_conn.Close()
			continue
		}
		go SwapConn(user_conn, home_conn)
	}
}


func SwapConn(conn1 net.Conn, conn2 net.Conn) {
	go Copy(conn1, conn2)
	Copy(conn2, conn1)
}

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