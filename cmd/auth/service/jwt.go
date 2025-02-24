package service

import (
	"TKMall/build/proto_gen/auth"
	"context"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)



func generateJWT(userId int32) (string, error) {
	claims := &jwt.StandardClaims{
		Subject:   fmt.Sprintf("%d", userId),
		ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(SecretKey))
}

func validateJWT(tokenString string) (*jwt.StandardClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*jwt.StandardClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("invalid token")
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

func (s *AuthServiceServer) TestGateWayMsg(ctx context.Context, req *auth.Empty) (*auth.TestGateWayMsgResp, error) {
	return &auth.TestGateWayMsgResp{Msg: "hello i am auth"}, nil
}
