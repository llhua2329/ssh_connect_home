package main

import (
	"net"
	"fmt"
	"flag"
	"strconv"
	"time"
)

func main() {
	ip := flag.String("ip", "127.0.0.1", "remote ip")
	port := flag.Int("port", 50001, "remote port")
	flag.Parse()

	server_addr := *ip +":"+ strconv.Itoa(*port)
	control_conn, err := net.Dial("tcp", server_addr)
	if err != nil {
		fmt.Println("connect server failed", err)
		return
	}
	defer control_conn.Close()

	var ctrl connection
	ctrl.conn = control_conn
	ctrl.recv = make(chan []byte)
	ctrl.send = make(chan []byte)
	go ctrl.Read()
	go ctrl.Write()

	recv := make([]byte, 1)
	tick := time.Tick(5 * time.Second)
	for {
		select {
		case <- tick: // 定时发送心跳
			heartbeat := []byte{0xFE}
			ctrl.send <- heartbeat
		case recv = <-ctrl.recv:
			if recv[0] == 0xFF { // 发起新的ssh连接
				fmt.Println(time.Now(), "accept now ssh")
				go newSsh(server_addr)
			}
		}
	}
}

func newSsh(server_info string) {
	fmt.Println(time.Now(), "send server to create new ssh")
	remote, err := net.Dial("tcp", server_info)
	if err != nil {
		fmt.Println("ssh info connect to server:", err)
		return
	}
	defer remote.Close()
	fmt.Println(time.Now(), "send local ssh")
	local, err := net.Dial("tcp", "127.0.0.1:22")
	if err != nil {
		fmt.Println("ssh info connect to 22:", err)
		return
	}
	fmt.Println(time.Now(), "swap data")
	defer local.Close()
	SwapConn(local, remote)
}
