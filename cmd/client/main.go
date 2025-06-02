package main

import (
	"context"
	"log"

	newsv1 "github.com/codeandlearn1991/news-grpc/api/news/v1"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("new client: %v\n", err)
	}

	client := newsv1.NewNewsServiceClient(conn)

	ctx := context.Background()

	createRes, err := client.Create(ctx, &newsv1.CreateRequest{
		Id:      uuid.NewString(),
		Author:  "Test Author",
		Title:   "Test title",
		Content: "Test content",
		Summary: "Test summary",
		Source:  "https://example.com",
		Tags:    []string{"tag1", "tag2"},
	})
	if err != nil {
		log.Fatalf("create news: %v", err)
	}

	getRes, err := client.Get(ctx, &newsv1.GetRequest{Id: createRes.Id})
	if err != nil {
		log.Fatalf("get news: %v", err)
	}
	log.Printf("get news: %+v", getRes)
}
