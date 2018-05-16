package main

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"time"

	"google.golang.org/grpc/examples/helloworld/helloworld"
	"gopkg.in/golang/protobuf.v1/proto"
)

/*
hand writter redis client for example,
you should use some real redis client library in production

redis> GRPC REQUEST greeter SayHello <binary pb payload>
 */
func main() {
	conn, err := net.Dial("tcp", "localhost:6380")
	if err != nil {
		panic(err)
	}

	conn.Write([]byte("*5\r\n$4\r\nGRPC\r\n$7\r\nREQUEST\r\n$7\r\ngreeter\r\n$8\r\nSayHello\r\n"))

	req := &helloworld.HelloRequest{}
	req.Name = fmt.Sprint(time.Now())

	payload, err := proto.Marshal(req)
	if err != nil {
		panic(err)
	}

	conn.Write([]byte(fmt.Sprintf("$%d\r\n", len(payload))))
	conn.Write(payload)
	conn.Write([]byte("\r\n"))

	response := make([]byte, 0)
	buf := make([]byte, 65535)
	resultLen := 0
	err = nil
	for err != io.EOF {
		bufSize, err := conn.Read(buf)
		if err != io.EOF && err != nil {
			panic(err)
		}

		if resultLen == 0 {
			if buf[0] != '$' {
				fmt.Printf("bad response: %s", string(buf))
				return
			}

			// consume $
			buf = buf[1:]
			bufSize--

			for {
				if len(buf) <= 0 {
					fmt.Printf("bad response: %s", string(buf))
					return
				}
				if buf[0] == '\r' && buf[1] == '\n' {
					buf = buf[2:]
					bufSize -= 2
					break
				}
				c, err := strconv.Atoi(string(buf[0]))
				if err != nil {
					panic(err)
				}
				resultLen = resultLen*10 + c
				buf = buf[1:]
				bufSize--
			}
		}

		response = append(response, buf[:bufSize]...)

		if resultLen > 0 && len(response) >= resultLen {
			break
		}
	}

	reply := &helloworld.HelloReply{}
	// slice -2 for \r\n
	err = proto.Unmarshal(response[:len(response)-2], reply)
	if err != nil {
		panic(err)
	}

	fmt.Printf("type=%T,message=%s", reply, reply.Message)
}
