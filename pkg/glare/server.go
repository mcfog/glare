package glare

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"reflect"
	"time"

	"github.com/secmask/go-redisproto"
	"gopkg.in/golang/protobuf.v1/proto"
	"gopkg.in/pkg/errors.v0"
)

type Server struct {
	handlerRequest func(argv [][]byte, payload []byte) ([]byte, error)
}

func (svr *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	parser := redisproto.NewParser(conn)
	writer := redisproto.NewWriter(bufio.NewWriter(conn))
	var ew error
	for {
		command, err := parser.ReadCommand()
		if err != nil {
			_, ok := err.(*redisproto.ProtocolError)
			if ok {
				ew = writer.WriteError(err.Error())
				continue
			} else {
				log.Println(err, " closed connection to ", conn.RemoteAddr())
				break
			}
		} else {
			ctx := &requestContext{
				writer:  writer,
				command: command,
			}
			switch {
			case ctx.MatchAndShift("GRPC"):
				ew = handleGrpc(ctx, svr)
			default:
				ew = writer.WriteError("Command not support")
			}
		}

		if command.IsLast() {
			writer.Flush()
		}
		if ew != nil {
			log.Println("Connection closed", ew)
			break
		}
	}
}

func (svr *Server) Run(network, address string) {
	listener, err := net.Listen(network, address)
	if err != nil {
		panic(err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error on accept: ", err)
			continue
		}
		go svr.handleConnection(conn)
	}
}

type ReflectServer struct {
	*Server
}

func NewReflectServer(clientMap map[string]interface{}, timeout time.Duration) *ReflectServer {
	svr := ReflectServer{
		Server: &Server{},
	}
	svr.handlerRequest = func(argv [][]byte, payload []byte) ([]byte, error) {
		if len(argv) != 2 {
			return nil, errors.New("bad param")
		}

		client, ok := clientMap[string(argv[0])]
		if !ok {
			return nil, errors.New("client not found")
		}
		methodName := string(argv[1])

		methodValue := reflect.ValueOf(client).MethodByName(methodName)
		if !methodValue.IsValid() {
			return nil, errors.New("method not found")
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		grpcArgc := methodValue.Type().NumIn() - 1
		grpcArgv := make([]reflect.Value, grpcArgc)
		for i := 0; i < grpcArgc; i++ {
			if i == 0 {
				grpcArgv[i] = reflect.ValueOf(ctx)
			} else {
				value := reflect.New(methodValue.Type().In(i).Elem())
				msg, ok := value.Interface().(proto.Message)
				if !ok {
					return nil, errors.New(fmt.Sprintf("failed to instantiate proto.Message %#v", value))
				}

				err := proto.Unmarshal(payload, msg)
				if err != nil {
					return nil, errors.Wrap(err, "failed to parse payload")
				}

				grpcArgv[i] = value
			}
		}
		callResult := methodValue.Call(grpcArgv)
		if !callResult[1].IsNil() {
			err, ok := callResult[1].Interface().(error)
			if ok {
				return nil, errors.Wrap(err, "bad invocation")
			} else {
				return nil, errors.New("bad invocation")
			}
		}

		msg, ok := callResult[0].Interface().(proto.Message)
		resultBytes, err := proto.Marshal(msg)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse response")
		}
		return resultBytes, nil
	}

	return &svr
}
