package services

import (
	"errors"
	"gorm.io/gorm"

	"github.com/chivta/spotiarch/internal/logger"
	"github.com/chivta/spotiarch/internal/models"
)
func NewUserService(db *gorm.DB, log *logger.Logger) *UserService {
	return &UserService{
		db:  db,
		log: log,
	}
}

type UserService struct {
	db  *gorm.DB
	log *logger.Logger
}

func (s *UserService) GetUserByID(userID string) (*models.User, error) {
	s.log.Infof("Fetching user by ID: %s", userID)

	var user models.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.log.Warnf("User not found: %s", userID)
			return nil, ErrUserNotFound
		}
		s.log.Errorf("Error fetching user: %v", err)
		return nil, ErrInternal
	}

	return &user, nil
}