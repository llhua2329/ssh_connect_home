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

	var ctrl connection
	ctrl.conn = control_conn
	ctrl.recv = make(chan []byte)
	ctrl.send = make(chan []byte)
	go ctrl.Read()
	go ctrl.Write()
	Run(ctrl, server_addr)
}

func Run(ctrl connection, server_addr string) {
	heartbeat := make([]byte, 1)
	heartbeat[0] = 0xFE
	byte := make([]byte, 1)
	tick := time.Tick(5 * time.Second)
	for {
		select {
		case <- tick: // 定时发送心跳
			ctrl.send <- heartbeat
		case byte = <-ctrl.recv:
			if byte[0] == 0xFF { // 发起新的ssh连接
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
	local, err := net.Dial("tcp", "192.168.1.200:22")
	if err != nil {
		fmt.Println("ssh info connect to 22:", err)
		return
	}
	fmt.Println(time.Now(), "swap data")
	defer local.Close()
	SwapConn(local, remote)
}