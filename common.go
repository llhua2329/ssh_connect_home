package main

import (
	"fmt"
	"net"
	"strconv"
)

type connection struct {
	conn net.Conn
	recv chan []byte
	send chan []byte
}

func (c connection) Read() {
	defer c.conn.Close()
	buf := make([]byte, 1024)
	for {
		len, err := c.conn.Read(buf)
		if err != nil {
			fmt.Println("read error ", err)
			return
		}
		if len == 0 {
			fmt.Println("read 0")
			return
		}
		c.recv <-buf[0:len]
	}
}

func (c connection) Write() {
	for {
		select {
		case buf := <-c.send:
			_, err := c.conn.Write(buf)
			if err != nil {
				fmt.Println("write err ", err)
				return
			}
		}
	}
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

func SwapConn(conn1 net.Conn, conn2 net.Conn) {
	go CopyConnection(conn1, conn2)
	CopyConnection(conn2, conn1)
}


type Acceptor struct {
	lister net.Listener
	conn chan net.Conn
}

func (l *Acceptor) Run(port int) error {
	lister, err := net.Listen("tcp", "0.0.0.0:" + strconv.Itoa(port))
	if err != nil {
		fmt.Println(err)
		return err
	}
	l.conn = make(chan net.Conn)
	l.lister = lister
	go l.accept()
	return nil
}

func (l *Acceptor) accept() {
	for {
		conn, err := l.lister.Accept()
		if err != nil {
			fmt.Println("accept err", err)
			break
		}
		l.conn <- conn
		fmt.Println("accept new connect ")
	}
	l.lister.Close()
}