package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	user "TKMall/build/proto_gen/user"

	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

type server struct {
	user.UnimplementedUserServiceServer
}

func (s *server) Login(_ context.Context, req *user.LoginReq) (*user.LoginResp, error) {
	log.Printf("Received: %v, %v", req.GetEmail(), req.GetPassword())
	return &user.LoginResp{UserId: 114514}, nil
}

func registerService(client *clientv3.Client, serviceName, serviceAddr string, ttl int64) error {
	leaseResp, err := client.Grant(context.Background(), ttl)
	if err != nil {
		return err
	}

	_, err = client.Put(context.Background(), serviceName, serviceAddr, clientv3.WithLease(leaseResp.ID))
	if err != nil {
		return err
	}

	ch, err := client.KeepAlive(context.Background(), leaseResp.ID)
	if err != nil {
		return err
	}

	go func() {
		for {
			<-ch
		}
	}()

	return nil
}

func main() {
	flag.Parse()

	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	err = registerService(client, "user-service", "localhost:50051", 10)
	if err != nil {
		log.Fatal(err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	user.RegisterUserServiceServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
