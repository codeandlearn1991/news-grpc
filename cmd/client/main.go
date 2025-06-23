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

func main() { //nolint:gocyclo // Refactor to reduce complexity
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

	getAllStream, err := client.GetAll(ctx, &emptypb.Empty{})
	if err != nil {
		log.Fatalf("get all news: %v", err)
	}

	allNews := make([]*newsv1.GetAllResponse, 0)

	for {
		getAllNews, getAllErr := getAllStream.Recv()
		if errors.Is(getAllErr, io.EOF) {
			break
		}

		if getAllErr != nil {
			log.Fatalf("get all news stream: %v", getAllErr)
		}

		allNews = append(allNews, getAllNews)
	}

	var clientStream grpc.ClientStreamingClient[newsv1.CreateRequest, emptypb.Empty]
	for i, n := range allNews {
		clientStream, err = client.UpdateNews(ctx)
		if err != nil {
			log.Fatalf("update news stream: %v", err)
		}

		err = clientStream.Send(&newsv1.CreateRequest{
			Id:      n.Id,
			Author:  n.Author + fmt.Sprintf(" updated %d", i),
			Title:   n.Title,
			Tags:    n.Tags,
			Summary: n.Summary,
			Content: n.Content,
			Source:  n.Source,
		})
		if err != nil {
			log.Fatalf("update news send: %v", err)
		}
	}

	if _, closeErr := clientStream.CloseAndRecv(); closeErr != nil {
		log.Fatalf("client stream close: %v", closeErr)
	}

	getAllStream, err = client.GetAll(ctx, &emptypb.Empty{})
	if err != nil {
		log.Fatalf("get all news: %v", err)
	}

	updatedNews := make([]*newsv1.GetAllResponse, 0)

	for {
		res, recvErr := getAllStream.Recv()
		if errors.Is(recvErr, io.EOF) {
			break
		}

		if recvErr != nil {
			log.Fatalf("get all news stream: %v", recvErr)
		}

		updatedNews = append(updatedNews, res)
	}

	log.Printf("updated news: %v", updatedNews)

	// Bidirectional stream to delete news
	deleteStream, err := client.DeletedNews(ctx)
	if err != nil {
		log.Fatalf("delete news stream: %v", err)
	}

	waitc := make(chan struct{})

	go func() {
		defer close(waitc)
		for _, news := range allNews {
			sendErr := deleteStream.Send(&newsv1.NewsID{Id: news.Id})
			if sendErr != nil {
				log.Fatalf("deleting news: %v", sendErr)
			}
		}
		if closeErr := deleteStream.CloseSend(); closeErr != nil {
			log.Fatalf("close and send: %v", closeErr)
		}
	}()

	for {
		_, recvErr := deleteStream.Recv()
		if errors.Is(recvErr, io.EOF) {
			log.Printf("delete stream ended: %v", recvErr)
			break
		}
		if recvErr != nil {
			log.Fatalf("delete stream: %v", recvErr)
		}
		log.Println("news deleted")
	}

	<-waitc

	getAllStream, err = client.GetAll(ctx, &emptypb.Empty{})
	if err != nil {
		log.Fatalf("get all news: %v", err)
	}

	allNews = make([]*newsv1.GetAllResponse, 0)

	for {
		getAllNews, getAllErr := getAllStream.Recv()
		if errors.Is(getAllErr, io.EOF) {
			break
		}

		if getAllErr != nil {
			log.Fatalf("get all news stream: %v", getAllErr)
		}

		allNews = append(allNews, getAllNews)
	}
	log.Println(allNews)
}
