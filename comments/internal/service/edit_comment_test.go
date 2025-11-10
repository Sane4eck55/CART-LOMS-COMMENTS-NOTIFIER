package service

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/comments/internal/model"
	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestHandler_CommentEdit(t *testing.T) {
	testRequest := model.Comment{
		// nolint:gosec
		ID: rand.Int63(),
		// nolint:gosec
		UserID: rand.Int63(),
		// nolint:gosec
		Sku:     rand.Int63(),
		Comment: "test comment",
	}

	tests := []struct {
		name               string
		testRequest        model.Comment
		setupMock          func(tc testComponent)
		expectedStatusCode codes.Code
		expectedErr        error
	}{
		{
			name:        "success",
			testRequest: testRequest,
			setupMock: func(tc testComponent) {
				tc.mock.GetCommentByIDMock.
					Expect(minimock.AnyContext, testRequest.ID).
					Return(&model.Comment{
						UserID:    testRequest.UserID,
						CreatedAt: time.Now(),
					}, nil)

				tc.mock.EditCommentMock.
					Expect(minimock.AnyContext, testRequest).
					Return(nil)
			},
			expectedStatusCode: codes.OK,
			expectedErr:        nil,
		},
		{
			name:        "комментарий не найден",
			testRequest: testRequest,
			setupMock: func(tc testComponent) {
				tc.mock.GetCommentByIDMock.
					Expect(minimock.AnyContext, testRequest.ID).
					Return(nil, model.ErrCommentNotFoundByID)

			},
			expectedStatusCode: codes.Internal,
			expectedErr:        model.ErrCommentNotFoundByID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := setupTest(t)
			tt.setupMock(tc)

			err := tc.service.CommentEdit(context.Background(), tt.testRequest)
			assert.Equal(t, tt.expectedErr, err)
			if st, ok := status.FromError(err); ok {
				assert.Equal(t, tt.expectedStatusCode, st.Code())
			}

		})
	}

}
