package server

import (
	"context"
	"errors"
	"math/rand"
	"testing"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/comments/internal/model"
	pb "github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/comments/pkg/api/v1"
	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestHandler_CommentListByUser(t *testing.T) {
	testRequest := &pb.CommentListByUserRequest{
		// nolint:gosec
		UserId: rand.Int63(),
	}
	testComments := []*pb.CommentByUserID{
		{
			// nolint:gosec
			Id:        rand.Int63(),
			Sku:       1,
			Comment:   "test_comment_1",
			CreatedAt: timestamppb.Now(),
		},
		{
			// nolint:gosec
			Id:        rand.Int63(),
			Sku:       2,
			Comment:   "test_comment_2",
			CreatedAt: timestamppb.Now(),
		},
	}
	expectResp := &pb.CommentListByUserResponse{
		Comments: testComments,
	}

	testModelComments := []model.Comment{
		{
			ID:        testComments[0].Id,
			Sku:       1,
			Comment:   "test_comment_1",
			CreatedAt: testComments[0].CreatedAt.AsTime(),
		},
		{
			ID:        testComments[1].Id,
			Sku:       2,
			Comment:   "test_comment_2",
			CreatedAt: testComments[1].CreatedAt.AsTime(),
		},
	}

	tests := []struct {
		name               string
		testRequest        *pb.CommentListByUserRequest
		setupMock          func(tc testComponent)
		expectedStatusCode codes.Code
		expectedResp       *pb.CommentListByUserResponse
		expectedErr        error
	}{
		{
			name:        "success",
			testRequest: testRequest,
			setupMock: func(tc testComponent) {
				tc.mock.CommentListByUserMock.
					Expect(minimock.AnyContext, testRequest.GetUserId()).
					Return(testModelComments, nil)
			},
			expectedStatusCode: codes.OK,
			expectedResp:       expectResp,
			expectedErr:        nil,
		},
		{
			name:        "error",
			testRequest: testRequest,
			setupMock: func(tc testComponent) {
				tc.mock.CommentListByUserMock.
					Expect(minimock.AnyContext, testRequest.GetUserId()).
					Return(nil, errors.New("test err"))

			},
			expectedStatusCode: codes.Internal,
			expectedResp:       nil,
			expectedErr:        errors.New("test err"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := setupTest(t)
			tt.setupMock(tc)

			resp, err := tc.server.CommentListByUser(context.Background(), tt.testRequest)
			assert.Equal(t, tt.expectedErr, err)
			assert.Equal(t, tt.expectedResp, resp)
			if st, ok := status.FromError(err); ok {
				assert.Equal(t, tt.expectedStatusCode, st.Code())
			}

		})
	}

}
