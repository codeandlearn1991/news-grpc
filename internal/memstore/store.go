package memstore

import (
	"net/url"
	"sync"
	"time"

	"github.com/google/uuid"
)

// News model used by store.
type News struct {
	// ID unique to the news.
	ID uuid.UUID
	// Author of the news.
	Author string
	// Title of the news.
	Title string
	// Summary of the news.
	Summary string
	// Content of the news.
	Content string
	// Source of the news.
	Source *url.URL
	// Tags associated with news.
	Tags []string
	// CreatedAt timestamp of the news.
	CreatedAt time.Time
	// UpdatedAt timestamp of the news.
	UpdatedAt time.Time
	// DeletedAt timestamp of the news.
	DeletedAt time.Time
}

// Store in-memory implementation.
type Store struct {
	lock sync.RWMutex
	news []*News
}

// New constructor for the store.
func New() *Store {
	return &Store{
		lock: sync.RWMutex{},
		news: make([]*News, 0),
	}
}

// Create news in the inmemory store.
func (s *Store) Create(news *News) *News {
	createdNews := &News{
		ID:        uuid.New(),
		Author:    news.Author,
		Title:     news.Title,
		Summary:   news.Summary,
		Content:   news.Content,
		Source:    news.Source,
		Tags:      news.Tags,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	s.news = append(s.news, createdNews)
	return createdNews
}

// Get news by it's id.
func (s *Store) Get(id uuid.UUID) *News {
	s.lock.RLock()
	defer s.lock.RUnlock()
	for _, news := range s.news {
		if news.ID == id && news.DeletedAt.IsZero() {
			return news
		}
	}
	return nil
}
