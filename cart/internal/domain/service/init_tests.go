package service

import (
	"testing"

	mock "github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/internal/domain/service/mocks"
	"github.com/gojuno/minimock/v3"
	"golang.org/x/time/rate" //go 1.23.4 go get golang.org/x/time@v0.10.0
)

type testServiceComponent struct {
	mockPS    *mock.ProductServiceMock
	mockRepo  *mock.RepositoryMock
	mockLoms  *mock.LomsMock
	mockTrace *mock.TracerMock
	service   *Service
}

func setupTest(t *testing.T) testServiceComponent {
	mc := minimock.NewController(t)
	mockPS := mock.NewProductServiceMock(mc)
	mockRepo := mock.NewRepositoryMock(mc)
	mockLoms := mock.NewLomsMock(mc)
	limiterPS := rate.NewLimiter(rate.Limit(10), 10)

	mockTrace := mock.NewTracerMock(mc)

	service := NewService(mockPS, mockRepo, mockLoms, limiterPS, mockTrace)

	return testServiceComponent{
		mockPS:    mockPS,
		mockRepo:  mockRepo,
		mockLoms:  mockLoms,
		mockTrace: mockTrace,
		service:   service,
	}
}
