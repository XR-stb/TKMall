package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"TKMall/build/proto_gen/auth"

	"github.com/dgrijalva/jwt-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	secretKey = "TikiTokMall"
)

type AuthServiceServer struct {
	auth.UnimplementedAuthServiceServer
}

func (s *AuthServiceServer) DeliverTokenByRPC(ctx context.Context, req *auth.DeliverTokenReq) (*auth.DeliveryResp, error) {
	token, err := generateJWT(req.UserId)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %v", err)
	}
	return &auth.DeliveryResp{Token: token}, nil
}

func (s *AuthServiceServer) VerifyTokenByRPC(ctx context.Context, req *auth.VerifyTokenReq) (*auth.VerifyResp, error) {
	_, err := validateJWT(req.Token)
	if err != nil {
		return &auth.VerifyResp{Res: false}, nil
	}
	return &auth.VerifyResp{Res: true}, nil
}

func generateJWT(userId int32) (string, error) {
	claims := &jwt.StandardClaims{
		Subject:   fmt.Sprintf("%d", userId),
		ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

func validateJWT(tokenString string) (*jwt.StandardClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*jwt.StandardClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("invalid token")
}

func main() {
	server := grpc.NewServer()
	auth.RegisterAuthServiceServer(server, &AuthServiceServer{})
	reflection.Register(server)

	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
