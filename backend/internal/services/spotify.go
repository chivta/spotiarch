package services

import (
	"gorm.io/gorm"
	
	"github.com/chivta/spotiarch/internal/logger"
)

func NewSpotifyService(db *gorm.DB, log *logger.Logger) *SpotifyService {
	return &SpotifyService{
		db:  db,
		log: log,
	}
}

type SpotifyService struct {
	db  *gorm.DB
	log *logger.Logger
}
