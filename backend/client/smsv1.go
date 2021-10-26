package client

//go:generate mockgen -package client -destination ./smsv1_client_mock.go -mock_names SmsServiceClient=MockSmsServiceClient github.com/sean-ahn/user/proto/gen/go/sms/v1 SmsServiceClient

import (
	"context"
	"sync"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	smsv1 "github.com/sean-ahn/user/proto/gen/go/sms/v1"
)

const smsServiceConfig = `{"loadBalancingPolicy":"round_robin"}`

var (
	smsOnce sync.Once
	smsCli  smsv1.SmsServiceClient

	_ smsv1.SmsServiceClient = (*MockSmsServiceClient)(nil)
)

func GetSmsV1Service(serviceHost string) smsv1.SmsServiceClient {
	smsOnce.Do(func() {
		conn, err := grpc.Dial(
			serviceHost,
			grpc.WithInsecure(),
			grpc.WithDefaultServiceConfig(smsServiceConfig),
		)
		if err != nil {
			logrus.Panic(err)
		}

		smsCli = smsv1.NewSmsServiceClient(conn)
	})

	return smsCli
}

type mockSmsV1ServiceClient struct{}

func (c *mockSmsV1ServiceClient) Send(_ context.Context, _ *smsv1.SendRequest, _ ...grpc.CallOption) (*smsv1.SendResponse, error) {
	return &smsv1.SendResponse{}, nil
}

func GetMockSmsV1Service(_ string) smsv1.SmsServiceClient {
	return &mockSmsV1ServiceClient{}
}
