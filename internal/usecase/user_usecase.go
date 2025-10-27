package usecase

import (
	"be-parkir/internal/domain/entities"
	"be-parkir/internal/repository"
	"errors"
)

type UserUsecase interface {
	GetProfile(userID uint) (*entities.User, error)
	UpdateProfile(userID uint, req *entities.UpdateUserRequest) (*entities.User, error)
	GetUserByID(userID uint) (*entities.User, error)
}

type userUsecase struct {
	userRepo repository.UserRepository
}

func NewUserUsecase(userRepo repository.UserRepository) UserUsecase {
	return &userUsecase{
		userRepo: userRepo,
	}
}

func (u *userUsecase) GetProfile(userID uint) (*entities.User, error) {
	user, err := u.userRepo.GetByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}
	// Remove jukir_profile from response to avoid circular reference
	user.JukirProfile = nil
	return user, nil
}

func (u *userUsecase) UpdateProfile(userID uint, req *entities.UpdateUserRequest) (*entities.User, error) {
	user, err := u.userRepo.GetByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Update fields if provided
	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.Email != nil {
		// Check if email is already taken by another user
		existingUser, err := u.userRepo.GetByEmail(*req.Email)
		if err == nil && existingUser != nil && existingUser.ID != userID {
			return nil, errors.New("email already taken")
		}
		user.Email = *req.Email
	}
	if req.Phone != nil {
		user.Phone = *req.Phone
	}

	if err := u.userRepo.Update(user); err != nil {
		return nil, errors.New("failed to update profile")
	}

	// Remove jukir_profile from response to avoid circular reference
	user.JukirProfile = nil
	return user, nil
}

func (u *userUsecase) GetUserByID(userID uint) (*entities.User, error) {
	user, err := u.userRepo.GetByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}
	// Remove jukir_profile from response to avoid circular reference
	user.JukirProfile = nil
	return user, nil
}
