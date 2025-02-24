package service

import (
	"TKMall/build/proto_gen/auth"
	"TKMall/build/proto_gen/user"

	"github.com/bwmarrin/snowflake"
	"gorm.io/gorm"
)

type User struct {
	ID       int32
	Email    string
	Password string
}

type UserServiceServer struct {
	user.UnimplementedUserServiceServer
	// Users      map[string]*User
	AuthClient auth.AuthServiceClient
	DB         *gorm.DB
	Node       *snowflake.Node
}

var userIDCounter int32 = 1
