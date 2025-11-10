package server

import (
	"testing"

	mock "github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/loms/internal/app/server/mocks"
	mockTracer "github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/loms/internal/service/mocks"
	"github.com/gojuno/minimock/v3"
)

type testComponent struct {
	mock       *mock.LomsServiceMock
	mockTracer *mockTracer.TracerMock
	server     *Server
}

func setupTest(t *testing.T) testComponent {
	mc := minimock.NewController(t)
	serviceMock := mock.NewLomsServiceMock(mc)
	tracerMock := mockTracer.NewTracerMock(mc)
	server := NewServer(serviceMock, tracerMock)

	return testComponent{
		mock:       serviceMock,
		mockTracer: tracerMock,
		server:     server,
	}
}
