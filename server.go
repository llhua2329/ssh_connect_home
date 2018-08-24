package main

import (
	"net"
	"fmt"
	"flag"
	"strconv"
	"time"
)

func main() {
	user_port := flag.Int("user_port", 50000, "user connet port")
	home_port := flag.Int("home_port", 50001, "home connect_port")
	flag.Parse()
	listen_user, err := net.Listen("tcp", "0.0.0.0:"+strconv.Itoa(*user_port))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer listen_user.Close()

	listen_home, err := net.Listen("tcp", "0.0.0.0:"+strconv.Itoa(*home_port))
	if err != nil {
		fmt.Println(time.Now(), err)
		return
	}
	defer listen_home.Close()

	control_conn, err := listen_home.Accept()
	if err != nil {
		fmt.Println(time.Now(), "control connect failed:", err)
		return
	}
	defer control_conn.Close()

	var ctrl connection
	ctrl.conn = control_conn
	ctrl.recv = make(chan []byte)
	ctrl.send = make(chan []byte)
	go ctrl.Read()
	go ctrl.Write()
	go func() { // 处理心跳
		buf := make([]byte, 1)
		heart := int64(time.Now().Unix())
		tick := time.Tick(3 * time.Second)
		for {
			select {
			case buf =<- ctrl.recv:
				if (buf[0] == 0xFE) {
					fmt.Println("heart from ctrl client")
					heart = int64(time.Now().Unix())
				}
			case <-tick:
				if int64(time.Now().Unix()) - heart > 20 {
					fmt.Println("heart timeout")
					ctrl.conn.Close()
					return
				}
			}
		}
	}()

	for {
		user_conn, err := listen_user.Accept()
		if err != nil {
			fmt.Println("accept err", err)
			continue
		}

		fmt.Println(time.Now(), "send new ssh notify to home client")
		b := make([]byte, 1)
		b[0] = 0xff
		ctrl.send <- b // 通知home 有新的ssh来连接

		fmt.Println(time.Now(), "wait for home_conn")
		home_conn, err := listen_home.Accept() // 等待home来连接
		if err != nil {
			user_conn.Close()
			fmt.Println(time.Now(), "accept home conn err:", err)
			continue
		}
		fmt.Println(time.Now(), "accept home connected:", home_conn.RemoteAddr())

		go SwapConn(user_conn, home_conn)
	}
}
