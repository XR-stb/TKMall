package service

import (
	"TKMall/build/proto_gen/user"

	"TKMall/common/proxy"
	"TKMall/common/events"

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
	DB       *gorm.DB
	Node     *snowflake.Node
	Proxy    proxy.ServiceProxy
	EventBus events.EventBus
}

var userIDCounter int32 = 1
