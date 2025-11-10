package server

import (
	"testing"

	mock "github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/comments/internal/app/server/mocks"
	"github.com/gojuno/minimock/v3"
)

type testComponent struct {
	mock   *mock.CommentsServiceMock
	server *Server
}

func setupTest(t *testing.T) testComponent {
	mc := minimock.NewController(t)
	serviceMock := mock.NewCommentsServiceMock(mc)
	server := NewServer(serviceMock)
	return testComponent{
		mock:   serviceMock,
		server: server,
	}
}
