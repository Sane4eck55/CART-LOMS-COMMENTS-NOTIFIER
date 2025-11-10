package service

import (
	"testing"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/loms/internal/service/mocks"
	"github.com/gojuno/minimock/v3"
)

type testComponent struct {
	mockRepo     *mocks.IRepositoryMock
	mockTracer   *mocks.TracerMock
	mockProducer *mocks.IProducerOrderEventMock
	service      *Service
}

func setupTest(t *testing.T) testComponent {
	mc := minimock.NewController(t)
	mockRepo := mocks.NewIRepositoryMock(mc)
	mockTracer := mocks.NewTracerMock(mc)
	mockProducer := mocks.NewIProducerOrderEventMock(mc)

	services := NewService(mockRepo, mockTracer, mockProducer)

	return testComponent{
		mockRepo:     mockRepo,
		mockTracer:   mockTracer,
		mockProducer: mockProducer,
		service:      &services,
	}
}
