package server

import (
	"context"
	"math/rand"
	"testing"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/comments/internal/model"
	pb "github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/comments/pkg/api/v1"
	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestHandler_CommentEdit(t *testing.T) {
	testRequest := &pb.CommentEditRequest{
		// nolint:gosec
		CommentId: rand.Int63(),
		// nolint:gosec
		UserId: rand.Int63(),

		NewComment: "test new comment",
	}

	expectResp := &pb.CommentEditResponse{}

	tests := []struct {
		name               string
		testRequest        *pb.CommentEditRequest
		setupMock          func(tc testComponent)
		expectedStatusCode codes.Code
		expectedResp       *pb.CommentEditResponse
		expectedErr        error
	}{
		{
			name:        "success",
			testRequest: testRequest,
			setupMock: func(tc testComponent) {
				tc.mock.CommentEditMock.
					Expect(minimock.AnyContext, convertToEditComment(testRequest)).
					Return(nil)
			},
			expectedStatusCode: codes.OK,
			expectedResp:       expectResp,
			expectedErr:        nil,
		},
		{
			name:        "комментарий не принадлежит этому пользователю",
			testRequest: testRequest,
			setupMock: func(tc testComponent) {
				tc.mock.CommentEditMock.
					Expect(minimock.AnyContext, convertToEditComment(testRequest)).
					Return(model.ErrCommentNotBelongUser)

			},
			expectedStatusCode: codes.PermissionDenied,
			expectedResp:       nil,
			expectedErr:        status.Error(codes.PermissionDenied, model.ErrCommentNotBelongUser.Error()),
		},
		{
			name:        "время для редактирования прошло",
			testRequest: testRequest,
			setupMock: func(tc testComponent) {
				tc.mock.CommentEditMock.
					Expect(minimock.AnyContext, convertToEditComment(testRequest)).
					Return(model.ErrTimeForEditOver)

			},
			expectedStatusCode: codes.InvalidArgument,
			expectedResp:       nil,
			expectedErr:        status.Error(codes.InvalidArgument, model.ErrTimeForEditOver.Error()),
		},
		{
			name:        "комментарий не найден",
			testRequest: testRequest,
			setupMock: func(tc testComponent) {
				tc.mock.CommentEditMock.
					Expect(minimock.AnyContext, convertToEditComment(testRequest)).
					Return(model.ErrCommentNotFoundByID)

			},
			expectedStatusCode: codes.NotFound,
			expectedResp:       nil,
			expectedErr:        status.Error(codes.NotFound, model.ErrCommentNotFoundByID.Error()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := setupTest(t)
			tt.setupMock(tc)

			resp, err := tc.server.CommentEdit(context.Background(), tt.testRequest)
			assert.Equal(t, tt.expectedErr, err)
			assert.Equal(t, tt.expectedResp, resp)
			if st, ok := status.FromError(err); ok {
				assert.Equal(t, tt.expectedStatusCode, st.Code())
			}

		})
	}

}
