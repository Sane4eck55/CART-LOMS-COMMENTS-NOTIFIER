package repository

import (
	"context"
	"math/rand"
	"testing"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/internal/domain/model"
	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/pkg/tracer"
)

// ServiceName ...
const ServiceName = "cart"

func BenchmarkInMemoryRepository_Add(b *testing.B) {
	ctx := context.Background()

	inputs := make([]model.RequestData, b.N)
	for i := 0; i < b.N; i++ {
		inputs[i] = model.RequestData{
			// nolint:gosec
			UserID: rand.Int63(),
			// nolint:gosec
			Sku: rand.Int63(),
			// nolint:gosec
			Count: rand.Uint32(),
		}
	}

	//url := "localhost:6831"
	t, err := tracer.NewTracer(ctx)
	if err != nil {
		b.Fatalf("NewTracer : %v", err)
	}

	repo := NewInMemoryRepository(t.Tracer)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := repo.Add(ctx, inputs[i])
		if err != nil {
			b.Fatalf("err add : %v", err)
		}
	}
}
