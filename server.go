package main

import (
	"net"
	"fmt"
	"flag"
	"strconv"
	"time"
	"os"
)

type myListener struct {
	conn chan net.Conn
}

func (l *myListener) Init(port int) {
	lister, err := net.Listen("tcp", "0.0.0.0:" + strconv.Itoa(port))
	if err != nil {
		fmt.Println(err)
		os.Exit(10)
	}
	defer lister.Close()
	for {
		conn, err := lister.Accept()
		if err != nil {
			fmt.Println("accept err", err)
			break
		}
		fmt.Println("accept new ")
		l.conn <- conn
		fmt.Println("accept new connect ")
	}
}

func main() {
	user_port := flag.Int("user_port", 50000, "user connet port")
	home_port := flag.Int("home_port", 50001, "home connect_port")
	flag.Parse()

	var home, user myListener
	home.conn = make(chan net.Conn)
	user.conn = make(chan net.Conn)
	go home.Init(*home_port)
	go user.Init(*user_port)

	var ctrl connection
	ctrl.recv = make(chan []byte)
	ctrl.send = make(chan []byte)

	ctrl.conn = <- home.conn
	fmt.Println("ctrl connection connected")
	go ctrl.Read()
	go ctrl.Write()

	heart := int64(time.Now().Unix())
	tick := time.Tick(3 * time.Second)
	recv := make([]byte ,1)
	for {
		select {
		case user_conn := <- user.conn: // 处理新连接
			go func() {
				fmt.Println(time.Now(), "send new ssh notify to home client")
				b := []byte{0xff}
				ctrl.send <- b // 通知home 有新的ssh来连接

				fmt.Println(time.Now(), "wait for home_conn")
				home_conn := <- home.conn // 等待home来连接
				fmt.Println(time.Now(), "accept home connected:", home_conn.RemoteAddr())
				go SwapConn(user_conn, home_conn)
			}()
		case recv =<- ctrl.recv:
			if (recv[0] == 0xFE) {
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
}
