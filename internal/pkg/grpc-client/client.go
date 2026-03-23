package grpcclient

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb_classification "healthfit-platform/proto/gen/ml_classification"
	pb_generation "healthfit-platform/proto/gen/ml_generation"
)

type MLClient struct {
	classification pb_classification.MLClassificationClient
	generation     pb_generation.MLGenerationClient
	conn           *grpc.ClientConn
	genConn        *grpc.ClientConn
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
			if classificationConn != nil {
				classificationConn.Close()
			}
			return nil, err
		}
	}

	return &MLClient{
		classification: pb_classification.NewMLClassificationClient(classificationConn),
		generation:     pb_generation.NewMLGenerationClient(generationConn),
		conn:           classificationConn,
		genConn:        generationConn,
	}, nil
}

func (c *MLClient) Classify(ctx context.Context, req *pb_classification.ClassifyRequest) (*pb_classification.ClassifyResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	return c.classification.Classify(ctx, req)
}

func (c *MLClient) GenerateProgram(ctx context.Context, req *pb_generation.GenerateRequest) (*pb_generation.GenerateResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	return c.generation.Generate(ctx, req)
}

func (c *MLClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
	if c.genConn != nil {
		c.genConn.Close()
	}
}