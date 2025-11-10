package service

import (
	"context"
	"errors"
	"math/rand"
	"testing"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/comments/internal/model"
	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestHandler_CommentAdd(t *testing.T) {
	testRequest := model.Comment{
		// nolint:gosec
		UserID: rand.Int63(),
		// nolint:gosec
		Sku:     rand.Int63(),
		Comment: "test comment",
	}
	// nolint:gosec
	expectResp := rand.Int63()

	tests := []struct {
		name               string
		testRequest        model.Comment
		setupMock          func(tc testComponent)
		expectedStatusCode codes.Code
		expectedResp       int64
		expectedErr        error
	}{
		{
			name:        "success",
			testRequest: testRequest,
			setupMock: func(tc testComponent) {
				tc.mock.AddCommentMock.
					Expect(minimock.AnyContext, testRequest).
					Return(expectResp, nil)
			},
			expectedStatusCode: codes.OK,
			expectedResp:       expectResp,
			expectedErr:        nil,
		},
		{
			name:        "error",
			testRequest: testRequest,
			setupMock: func(tc testComponent) {
				tc.mock.AddCommentMock.
					Expect(minimock.AnyContext, testRequest).
					Return(model.ErrCommentID, errors.New("test err"))

			},
			expectedStatusCode: codes.Internal,
			expectedResp:       model.ErrCommentID,
			expectedErr:        errors.New("test err"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := setupTest(t)
			tt.setupMock(tc)

			resp, err := tc.service.CommentAdd(context.Background(), tt.testRequest)
			assert.Equal(t, tt.expectedErr, err)
			assert.Equal(t, tt.expectedResp, resp)
			if st, ok := status.FromError(err); ok {
				assert.Equal(t, tt.expectedStatusCode, st.Code())
			}

		})
	}

}
