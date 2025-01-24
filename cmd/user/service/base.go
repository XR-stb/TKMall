package service

import (
	"TKMall/build/proto_gen/auth"
	"TKMall/build/proto_gen/user"
)

type User struct {
	ID       int32
	Email    string
	Password string
}

type UserServiceServer struct {
	user.UnimplementedUserServiceServer
	Users      map[string]*User
	AuthClient auth.AuthServiceClient
}

var userIDCounter int32 = 1
