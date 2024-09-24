package service

import (
	"database/sql"
)

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	s := &Service{
		db: db,
	}
	s.Crawl()
	return s
}

// Crawl triggers the crawl services to run fetching tenders in the background
func (s *Service) Crawl() {
	// crawls the tenders we can get via rss
	go s.processFeeds()
}
