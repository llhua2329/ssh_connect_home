package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

func main() {
	ip := flag.String("ip", "127.0.0.1", "remote ip")
	remote_port := flag.Int("remote_port", 50001, "remote port")
	local_port := flag.Int("local_port", 22, "local_port")
	flag.Parse()

	server_addr := *ip + ":" + strconv.Itoa(*remote_port)
	local_addr := "192.168.1.200" + ":" + strconv.Itoa(*local_port)
	control_conn, err := net.Dial("tcp", server_addr)
	if err != nil {
		fmt.Println("connect server failed", err)
		return
	}
	defer control_conn.Close()

	ctrl := NewConnection()
	ctrl.conn = control_conn
	go ctrl.Read()
	go ctrl.Write()

	recv := make([]byte, 1)
	tick := time.Tick(5 * time.Second)
	for {
		select {
		case <-tick: // 定时发送心跳
			heartbeat := []byte{0xFE}
			ctrl.send <- heartbeat
		case recv = <-ctrl.recv:
			if recv[0] == 0xFF { // 发起新的ssh连接
				fmt.Println(time.Now(), "accept now ssh")
				go NewChannel(server_addr, local_addr)
			}
		case <-ctrl.read_close:
			fmt.Println("ctrl connect close")
			os.Exit(0)
		}
	}
}

func NewChannel(server_info string, local_addr string) {
	fmt.Println(time.Now(), "send server to create new ssh")
	remote, err := net.Dial("tcp", server_info)
	if err != nil {
		fmt.Println("ssh info connect to server:", err)
		return
	}
	defer remote.Close()
	fmt.Println(time.Now(), "send local ssh")
	local, err := net.Dial("tcp", local_addr)
	if err != nil {
		fmt.Println("ssh info connect to 22:", err)
		return
	}
	fmt.Println(time.Now(), "swap data")
	defer local.Close()
	SwapConn(local, remote)
}
