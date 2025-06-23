package grpc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"

	newsv1 "github.com/codeandlearn1991/news-grpc/api/news/v1"
	"github.com/codeandlearn1991/news-grpc/internal/memstore"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// NewsStorer to store news.
type NewsStorer interface {
	Create(news *memstore.News) *memstore.News
	Get(id uuid.UUID) *memstore.News
	GetAll() []*memstore.News
	Update(news *memstore.News)
	Delete(id uuid.UUID)
}

// Server implements of NewServiceServer.
type Server struct {
	newsv1.UnimplementedNewsServiceServer
	store NewsStorer
}

// NewServer returns an intialized instance of Server.
func NewServer(store NewsStorer) *Server {
	return &Server{
		store: store,
	}
}

// Create method implementation for the news gRPC server.
func (s *Server) Create(_ context.Context, in *newsv1.CreateRequest) (*newsv1.CreateResponse, error) {
	parsedNews, err := parseAndValidate(in)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	createdNews := s.store.Create(parsedNews)
	return toNewsResponse(createdNews), nil
}

// Get method implementation for the news gRPC server.
func (s *Server) Get(_ context.Context, in *newsv1.GetRequest) (*newsv1.GetResponse, error) {
	newsUUID, err := uuid.Parse(in.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	fetchedNews := s.store.Get(newsUUID)
	if fetchedNews == nil {
		return nil, status.Error(codes.NotFound, "news with given not found")
	}

	return &newsv1.GetResponse{
		Id:        fetchedNews.ID.String(),
		Author:    fetchedNews.Author,
		Title:     fetchedNews.Title,
		Summary:   fetchedNews.Summary,
		Content:   fetchedNews.Content,
		Source:    fetchedNews.Source.String(),
		Tags:      fetchedNews.Tags,
		CreatedAt: timestamppb.New(fetchedNews.CreatedAt.UTC()),
		UpdatedAt: timestamppb.New(fetchedNews.UpdatedAt.UTC()),
	}, nil
}

// GetAll news.
func (s *Server) GetAll(_ *emptypb.Empty, stream newsv1.NewsService_GetAllServer) error {
	for _, fetchedNews := range s.store.GetAll() {
		if err := stream.Send(&newsv1.GetAllResponse{
			Id:        fetchedNews.ID.String(),
			Author:    fetchedNews.Author,
			Title:     fetchedNews.Title,
			Summary:   fetchedNews.Summary,
			Content:   fetchedNews.Content,
			Source:    fetchedNews.Source.String(),
			Tags:      fetchedNews.Tags,
			CreatedAt: timestamppb.New(fetchedNews.CreatedAt.UTC()),
			UpdatedAt: timestamppb.New(fetchedNews.UpdatedAt.UTC()),
		}); err != nil {
			return err
		}
	}
	return nil
}

// UpdateNews gRPC method.
func (s *Server) UpdateNews(stream newsv1.NewsService_UpdateNewsServer) error {
	for {
		req, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return stream.SendAndClose(&emptypb.Empty{})
		}
		if err != nil {
			return err
		}
		updatedNews, err := parseAndValidate(req)
		if err != nil {
			return status.Errorf(codes.InvalidArgument, "validation failed %v", err)
		}
		s.store.Update(updatedNews)
	}
}

// DeletedNews from store.
func (s *Server) DeletedNews(stream newsv1.NewsService_DeletedNewsServer) error {
	for {
		req, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return nil
		}

		if err != nil {
			return err
		}

		newsUUID, err := uuid.Parse(req.Id)
		if err != nil {
			return status.Error(codes.InvalidArgument, err.Error())
		}

		s.store.Delete(newsUUID)
		if err := stream.Send(&emptypb.Empty{}); err != nil {
			return err
		}
	}
}

func parseAndValidate(in *newsv1.CreateRequest) (n *memstore.News, errs error) {
	if in == nil {
		return nil, errors.New("news request empty")
	}

	if in.Author == "" {
		errs = errors.Join(errs, errors.New("author cannot be empty"))
	}

	if in.Title == "" {
		errs = errors.Join(errs, errors.New("title cannot be empty"))
	}

	if in.Summary == "" {
		errs = errors.Join(errs, errors.New("summary cannot be empty"))
	}

	if in.Content == "" {
		errs = errors.Join(errs, errors.New("content cannot be empty"))
	}

	if in.Tags == nil {
		errs = errors.Join(errs, errors.New("tags cannot be nil"))
	}

	if len(in.Tags) == 0 {
		errs = errors.Join(errs, errors.New("tags cannot be empty"))
	}

	parsedID, err := uuid.Parse(in.Id)
	if err != nil {
		errs = errors.Join(errs, fmt.Errorf("invalid id: %w", err))
	}

	parsedURL, err := url.Parse(in.Source)
	if err != nil {
		errs = errors.Join(errs, fmt.Errorf("invalid url: %w", err))
	}

	if errs != nil {
		return nil, errs
	}

	return &memstore.News{
		ID:      parsedID,
		Author:  in.Author,
		Title:   in.Title,
		Summary: in.Summary,
		Content: in.Content,
		Source:  parsedURL,
		Tags:    in.Tags,
	}, nil
}

func toNewsResponse(news *memstore.News) *newsv1.CreateResponse {
	return &newsv1.CreateResponse{
		Id:        news.ID.String(),
		Author:    news.Author,
		Title:     news.Title,
		Summary:   news.Summary,
		Content:   news.Content,
		Source:    news.Source.String(),
		Tags:      news.Tags,
		CreatedAt: timestamppb.New(news.CreatedAt.UTC()),
		UpdatedAt: timestamppb.New(news.UpdatedAt.UTC()),
	}
}
