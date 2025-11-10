package service

import (
	"context"
	"errors"
	"math/rand"
	"testing"
	"time"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/comments/internal/model"
	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestHandler_CommentListByUser(t *testing.T) {
	// nolint:gosec
	testRequest := rand.Int63()

	expectedResp := []model.Comment{
		{
			// nolint:gosec
			ID:     rand.Int63(),
			UserID: testRequest,
			// nolint:gosec
			Sku:       rand.Int63(),
			Comment:   "test comment1",
			CreatedAt: time.Now(),
		},
		{
			// nolint:gosec
			ID: rand.Int63(),
			// nolint:gosec
			UserID: testRequest,
			// nolint:gosec
			Sku:       rand.Int63(),
			Comment:   "test comment2",
			CreatedAt: time.Now(),
		},
	}

	tests := []struct {
		name               string
		testRequest        int64
		setupMock          func(tc testComponent)
		expectedResp       []model.Comment
		expectedStatusCode codes.Code
		expectedErr        error
	}{
		{
			name:        "success",
			testRequest: testRequest,
			setupMock: func(tc testComponent) {
				tc.mock.CommentListByUserMock.
					Expect(minimock.AnyContext, testRequest).
					Return(expectedResp, nil)
			},
			expectedStatusCode: codes.OK,
			expectedErr:        nil,
		},
		{
			name:        "error",
			testRequest: testRequest,
			setupMock: func(tc testComponent) {
				tc.mock.CommentListByUserMock.
					Expect(minimock.AnyContext, testRequest).
					Return(nil, errors.New("test error"))

			},
			expectedStatusCode: codes.Internal,
			expectedErr:        errors.New("test error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := setupTest(t)
			tt.setupMock(tc)

			_, err := tc.service.CommentListByUser(context.Background(), tt.testRequest)
			assert.Equal(t, tt.expectedErr, err)
			if st, ok := status.FromError(err); ok {
				assert.Equal(t, tt.expectedStatusCode, st.Code())
			}

		})
	}

}
