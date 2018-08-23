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
	control_conn.SetDeadline(time.Time{})
	if err != nil {
		fmt.Println(time.Now(), "control connect failed:", err)
		return
	}
	defer control_conn.Close()
	control_write := make(chan []byte, 1)
	go CtrlWrite(control_write, control_conn)
	for {
		user_conn, err := listen_user.Accept()
		if err != nil {
			continue
		}

		fmt.Println(time.Now(), "send new ssh notify to home client")
		b := make([]byte, 1)
		b[0] = 0xff
		control_write <- b // 通知home 有新的ssh来连接

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

func CtrlWrite(w chan []byte, conn net.Conn) {
	tick := time.Tick(5 * time.Second)
	for {
		select {
		case buf := <-w:
			fmt.Println("send ", buf)
			conn.Write(buf)
		case <- tick:
			var buf [1]byte
			buf[0] = 0xFE
			w <- buf[:1]
		}
	}
}

func SwapConn(conn1 net.Conn, conn2 net.Conn) {
	conn1.SetDeadline(time.Time{})
	conn2.SetDeadline(time.Time{})
	go CopyConnection(conn1, conn2)
	CopyConnection(conn2, conn1)
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
