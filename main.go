package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"

	"github.com/mailgun/holster"
	"github.com/soheilhy/cmux"
	"github.com/thrawn01/grpc-http-1/pb"
	"google.golang.org/grpc"
)

type Server struct {
}

type ServerConfig struct {
	ListenAddress string
}

func (s *Server) Echo(ctx context.Context, req *pb.EchoRequest) (*pb.EchoResponse, error) {
	return &pb.EchoResponse{Message: req.Message}, nil
}

func checkErr(err error) {
	if err != nil {
		fmt.Printf("err: %s\n", err)
		os.Exit(1)
	}
}

func main() {

	if len(os.Args) == 1 {
		runServer()
		os.Exit(0)
	}

	if os.Args[1] == "http" {
		httpClient()
		return
	}
	grpcClient()
}

func httpClient() {
	resp, err := http.Get("http://localhost:8080")
	checkErr(err)

	body, err := ioutil.ReadAll(resp.Body)
	fmt.Printf("HTTP Resp: %s", string(body))
}

func grpcClient() {
	conn, err := grpc.Dial("localhost:8080", grpc.WithInsecure())
	checkErr(err)

	client := pb.NewEchoServiceClient(conn)

	resp, err := client.Echo(context.Background(), &pb.EchoRequest{Message: "hello"})
	checkErr(err)
	fmt.Printf("Resp: %s\n", resp.Message)
}

func runServer() {
	listener, err := net.Listen("tcp", "localhost:8080")
	checkErr(err)

	m := cmux.New(listener)
	//grpcListener := m.Match(cmux.HTTP2HeaderField("content-type", "application/grpc"))
	grpcListener := m.Match(cmux.HTTP2())
	httpListener := m.Match(cmux.HTTP1Fast())
	anyListener := m.Match(cmux.Any())

	var wg holster.WaitGroup

	wg.Run(func(interface{}) error {
		fmt.Print("Any Listening....\n")
		for {
			conn, err := anyListener.Accept()
			checkErr(err)

			fmt.Printf("Got Any\n")
			_, err = conn.Write([]byte("Hello"))
			checkErr(err)

			conn.Close()
		}
		return nil
	}, nil)

	wg.Run(func(interface{}) error {
		server := grpc.NewServer()
		s := Server{}
		pb.RegisterEchoServiceServer(server, &s)
		fmt.Print("GRPC Listening....\n")
		err := server.Serve(grpcListener)
		fmt.Print("GRPC Done\n")
		return err
	}, nil)

	wg.Run(func(interface{}) error {
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
			w.Write([]byte("pong"))
		})
		s := &http.Server{Handler: mux}
		fmt.Print("HTTP Listening....\n")
		err := s.Serve(httpListener)
		fmt.Print("HTTP Done\n")
		return err
	}, nil)

	wg.Run(func(interface{}) error {
		fmt.Printf("cmux.Server()....\n")
		return m.Serve()
	}, nil)

	fmt.Printf("Waiting....\n")
	if errs := wg.Wait(); err != nil {
		for _, err := range errs {
			fmt.Printf("serve err: %s", err)
		}
		os.Exit(1)
	}
}
