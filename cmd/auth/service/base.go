package service

import "TKMall/build/proto_gen/auth"

type AuthServiceServer struct {
	auth.UnimplementedAuthServiceServer
}

var SecretKey string