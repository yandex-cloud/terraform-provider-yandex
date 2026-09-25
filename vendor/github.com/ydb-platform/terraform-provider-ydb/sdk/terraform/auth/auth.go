package auth

import (
	"context"

	"github.com/ydb-platform/ydb-go-genproto/Ydb_Auth_V1"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Auth"
	"google.golang.org/grpc"
)

type YdbCredentials struct {
	Token    string
	User     string
	Password string
}

type GetAuthCallback func(ctx context.Context) (YdbCredentials, error)

func GetTokenFromStaticCreds(ctx context.Context, user, password string, conn *grpc.ClientConn) (string, error) {
	request := &Ydb_Auth.LoginRequest{
		User:     user,
		Password: password,
	}
	result := &Ydb_Auth.LoginResult{}

	stub := Ydb_Auth_V1.NewAuthServiceClient(conn)

	opResp, err := stub.Login(ctx, request)
	if err != nil {
		return "", nil
	}
	err = opResp.Operation.Result.UnmarshalTo(result)
	if err != nil {
		return "", nil
	}
	return result.Token, nil
}
