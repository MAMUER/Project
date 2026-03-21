package grpcclient

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type MLClient struct {
	classification pb.MLClassificationClient
	generation     pb.MLGenerationClient
	conn           *grpc.ClientConn
}

func NewMLClient(classificationAddr, generationAddr string) (*MLClient, error) {
	var classificationConn, generationConn *grpc.ClientConn
	var err error

	if classificationAddr != "" {
		classificationConn, err = grpc.Dial(classificationAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, err
		}
	}

	if generationAddr != "" {
		generationConn, err = grpc.Dial(generationAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			classificationConn.Close()
			return nil, err
		}
	}

	return &MLClient{
		classification: pb.NewMLClassificationClient(classificationConn),
		generation:     pb.NewMLGenerationClient(generationConn),
		conn:           classificationConn,
	}, nil
}

func (c *MLClient) Classify(ctx context.Context, req *pb.ClassifyRequest) (*pb.ClassifyResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	return c.classification.Classify(ctx, req)
}

func (c *MLClient) GenerateProgram(ctx context.Context, req *pb.GenerateRequest) (*pb.GenerateResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	return c.generation.Generate(ctx, req)
}

func (c *MLClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}