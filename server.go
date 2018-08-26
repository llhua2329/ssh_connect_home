package main

import (
	"fmt"
	"flag"
	"time"
)

func main() {
	user_port := flag.Int("user_port", 50000, "user connet port")
	home_port := flag.Int("home_port", 50001, "home connect_port")
	flag.Parse()

	var home, user Acceptor
	err := home.Run(*home_port)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = user.Run(*user_port)
	if err != nil {
		fmt.Println(err)
		return
	}

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
