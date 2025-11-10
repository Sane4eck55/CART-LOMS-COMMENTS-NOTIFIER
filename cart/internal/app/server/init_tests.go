package server

import (
	"testing"

	mock "github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/internal/app/server/mocks"
	mockTracer "github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/internal/domain/service/mocks"
	"github.com/gojuno/minimock/v3"
)

type testComponent struct {
	mock   *mock.ServiceMock
	server *Server
	tracer *mockTracer.TracerMock
}

func setupTest(t *testing.T) testComponent {
	mc := minimock.NewController(t)
	tracer := mockTracer.NewTracerMock(mc)
	serviceMock := mock.NewServiceMock(mc)
	server := NewServer(serviceMock, tracer)

	return testComponent{
		mock:   serviceMock,
		server: server,
		tracer: tracer,
	}
}
