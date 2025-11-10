// Package loms ...
package loms

import (
	"context"

	pb "github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/pkg/github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/loms/api/v1"
)

// Client ...
type Client struct {
	client pb.LomsClient
}

// NewLomsCliemt ...
func NewLomsCliemt(client pb.LomsClient) *Client {
	return &Client{
		client: client,
	}
}

// CreateOrder ...
func (c *Client) CreateOrder(ctx context.Context, req *pb.OrderCreateRequest) (*pb.OrderCreateResponse, error) {
	resp, err := c.client.OrderCreate(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// GetStocksInfo ...
func (c *Client) GetStocksInfo(ctx context.Context, req *pb.StocksInfoRequest) (*pb.StocksInfoResponse, error) {
	resp, err := c.client.StocksInfo(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
