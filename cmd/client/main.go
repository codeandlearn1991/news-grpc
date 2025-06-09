package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"

	newsv1 "github.com/codeandlearn1991/news-grpc/api/news/v1"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

func main() {
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("new client: %v\n", err)
	}

	client := newsv1.NewNewsServiceClient(conn)

	ctx := context.Background()

	for i := range 5 {
		_, err = client.Create(ctx, &newsv1.CreateRequest{
			Id:      uuid.NewString(),
			Author:  fmt.Sprintf("Test Author %d", i),
			Title:   fmt.Sprintf("Test title %d", i),
			Content: fmt.Sprintf("Test content %d", i),
			Summary: fmt.Sprintf("Test summary %d", i),
			Source:  "https://example.com",
			Tags:    []string{"tag1", "tag2"},
		})
		if err != nil {
			log.Fatalf("create news: %v", err)
		}
	}

	getAllRes, err := client.GetAll(ctx, &emptypb.Empty{})
	if err != nil {
		log.Fatalf("get all news: %v", err)
	}

	allNews := make([]*newsv1.GetAllResponse, 0)

	for {
		res, err := getAllRes.Recv()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			log.Fatalf("get all news stream: %v", err)
		}

		allNews = append(allNews, res)
	}

	log.Printf("all news: %v", allNews)
}
