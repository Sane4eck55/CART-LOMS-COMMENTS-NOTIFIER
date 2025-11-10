// Package app ...
package app

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"time"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/internal/app/server"
	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/internal/client/loms"
	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/internal/domain/model"
	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/internal/domain/repository"
	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/internal/domain/service"
	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/internal/infra/config"
	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/pkg/logger"
	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/pkg/tracer"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	product_service "github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/internal/domain/product-service"
	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/internal/infra/http/middlewares"
	retryclient "github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/internal/infra/http/retry_client"
	rt "github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/internal/infra/http/round_trippers"

	pbLoms "github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/pkg/github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/loms/api/v1"
)

const (
	// ServiceName ...
	ServiceName = "cart"
)

// App ...
type App struct {
	config   *config.Config
	server   http.Server
	service  *service.Service
	connLoms *grpc.ClientConn
	tracer   *tracer.TManager
}

// NewApp ...
func NewApp(ctx context.Context) (*App, error) {
	ctx = metadata.AppendToOutgoingContext(ctx, "service-name", ServiceName)

	logger.GetLogger(ServiceName)

	c, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("config.LoadConfig: %w", err)
	}

	app := &App{config: c}

	app.server.Handler, err = app.bootstrapHandlers(ctx)
	if err != nil {
		return nil, err
	}

	return app, nil
}

// ListenAndServe ...
func (app *App) ListenAndServe() error {
	address := fmt.Sprintf("%s:%s", app.config.Server.Host, app.config.Server.Port)

	l, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	logger.Infow(fmt.Sprintf("app bootstrap : %s", address))

	return app.server.Serve(l)
}

func (app *App) bootstrapHandlers(ctx context.Context) (http.Handler, error) {
	t, err := tracer.NewTracer(ctx)
	if err != nil {
		logger.Fatalw(fmt.Sprintf("NewTracer : %v", err))
	}
	app.tracer = t

	transport := http.DefaultTransport
	transport = rt.NewLogRoundTripper(transport)
	httpClient := http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}
	psClientConfig := retryclient.NewRetryClient(
		&httpClient,
		3,
		5*time.Second,
	)
	productService := product_service.NewProductService(
		*psClientConfig,
		app.config.ProductService.Token,
		fmt.Sprintf("%s:%s", app.config.ProductService.Host, app.config.ProductService.Port),
	)

	repo := repository.NewInMemoryRepository(t.Tracer)
	clientLoms, err := initClientLoms(app)
	if err != nil {
		return nil, fmt.Errorf("initClientLoms : %v", err)
	}

	limiterPS := rate.NewLimiter(rate.Limit(app.config.ProductService.Limit), app.config.ProductService.Burst)

	service := service.NewService(productService, repo, clientLoms, limiterPS, t.Tracer)

	s := server.NewServer(service, t.Tracer)

	mx := http.NewServeMux()
	mx.HandleFunc(model.AddItemURL, s.AddItem)
	mx.HandleFunc(model.DeleteItemURL, s.DeleteItem)
	mx.HandleFunc(model.DeleteItemsByUserIDURL, s.DeleteItemsByUserID)
	mx.HandleFunc(model.GetItemsByUserIDURL, s.GetItemsByUserID)
	mx.HandleFunc(model.OrderFullCartURL, s.OrderFullCart)
	mx.Handle(model.GetMetricsURL, promhttp.Handler())
	mx.HandleFunc(model.DebugPprof, pprofhandler)

	h := middlewares.NewTimerMiddleware(mx)

	return h, nil
}

// Close ...
func (app *App) Close(ctx context.Context) error {
	//nolint:errcheck, gosec
	app.connLoms.Close()
	logger.Infow("connect loms closed")
	app.service.Repository.Close()
	logger.Infow("connect repo closed")
	//nolint:errcheck, gosec
	app.tracer.TracerProvider.Shutdown(ctx)
	logger.Infow("tracer Shutdown")
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return app.server.Shutdown(ctx)

}

// lsof -iTCP:50051 -sTCP:LISTEN
func initClientLoms(app *App) (*loms.Client, error) {
	config := app.config
	url := fmt.Sprintf("%s:%s", config.LomsService.Host, config.LomsService.Port)
	conn, err := grpc.NewClient(
		url,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(middlewares.UnaryInterceptor),
	)
	if err != nil {
		return nil, err
	}

	app.connLoms = conn

	cliLoms := loms.NewLomsCliemt(pbLoms.NewLomsClient(conn))
	logger.Infow(fmt.Sprintf("Loms service at : %s", url))
	return cliLoms, nil
}

// pprofhandler ...
func pprofhandler(w http.ResponseWriter, r *http.Request) {
	http.DefaultServeMux.ServeHTTP(w, r)
}
