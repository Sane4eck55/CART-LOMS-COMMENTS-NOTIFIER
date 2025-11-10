package service

import (
	"testing"

	mock "github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/comments/internal/service/mocks"
	"github.com/gojuno/minimock/v3"
)

type testComponent struct {
	mock    *mock.IRepositoryMock
	service *Service
}

func setupTest(t *testing.T) testComponent {
	mc := minimock.NewController(t)
	repoMock := mock.NewIRepositoryMock(mc)
	service := NewService(repoMock)
	return testComponent{
		mock:    repoMock,
		service: service,
	}
}
