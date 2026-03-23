package integration

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "healthfit-platform/proto/gen/java_legacy_bridge"
)

type JavaLegacyClient struct {
	conn   *grpc.ClientConn
	client pb.JavaLegacyBridgeClient
}

func NewJavaLegacyClient(addr string) (*JavaLegacyClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Java legacy service: %w", err)
	}

	return &JavaLegacyClient{
		conn:   conn,
		client: pb.NewJavaLegacyBridgeClient(conn),
	}, nil
}

func (c *JavaLegacyClient) GetClubInfo(ctx context.Context, clubID string) (*pb.GetClubInfoResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return c.client.GetClubInfo(ctx, &pb.GetClubInfoRequest{ClubId: clubID})
}

func (c *JavaLegacyClient) GetEquipment(ctx context.Context, clubID string) ([]*pb.Equipment, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	resp, err := c.client.GetEquipment(ctx, &pb.GetEquipmentRequest{ClubId: clubID})
	if err != nil {
		return nil, err
	}
	return resp.Items, nil
}

func (c *JavaLegacyClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
