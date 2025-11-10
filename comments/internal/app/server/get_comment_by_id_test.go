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

func TestHandler_CommentGetByID(t *testing.T) {
	testRequest := &pb.CommentGetByIDRequest{
		// nolint:gosec
		Id: rand.Int63(),
	}

	expectResp := &pb.CommentGetByIDResponse{
		Id:        testRequest.GetId(),
		UserId:    1,
		Sku:       1,
		Comment:   "test_comment",
		CreatedAt: timestamppb.Now(),
	}

	tests := []struct {
		name               string
		testRequest        *pb.CommentGetByIDRequest
		setupMock          func(tc testComponent)
		expectedStatusCode codes.Code
		expectedResp       *pb.CommentGetByIDResponse
		expectedErr        error
	}{
		{
			name:        "success",
			testRequest: testRequest,
			setupMock: func(tc testComponent) {
				tc.mock.CommentGetByIDMock.
					Expect(minimock.AnyContext, testRequest.GetId()).
					Return(&model.Comment{
						ID:        expectResp.Id,
						UserID:    expectResp.UserId,
						Sku:       expectResp.Sku,
						Comment:   expectResp.Comment,
						CreatedAt: expectResp.CreatedAt.AsTime(),
					}, nil)
			},
			expectedStatusCode: codes.OK,
			expectedResp:       expectResp,
			expectedErr:        nil,
		},
		{
			name:        "error",
			testRequest: testRequest,
			setupMock: func(tc testComponent) {
				tc.mock.CommentGetByIDMock.
					Expect(minimock.AnyContext, testRequest.GetId()).
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

			resp, err := tc.server.CommentGetByID(context.Background(), tt.testRequest)
			assert.Equal(t, tt.expectedErr, err)
			assert.Equal(t, tt.expectedResp, resp)
			if st, ok := status.FromError(err); ok {
				assert.Equal(t, tt.expectedStatusCode, st.Code())
			}

		})
	}

}
