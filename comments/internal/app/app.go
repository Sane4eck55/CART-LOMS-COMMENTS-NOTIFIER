// Package app ...
package app

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/comments/configs"
	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/comments/internal/app/server"
	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/comments/internal/middlewares"
	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/comments/internal/repository"
	shardmanager "github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/comments/internal/repository/shard_manager"
	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/comments/internal/service"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"

	pb "github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/comments/pkg/api/v1"
)

const (
	// ServiceName ...
	ServiceName = "comments"
	// readHeaderTimeout ...
	readHeaderTimeout = 1 * time.Minute
	// numNodes ...
	numNodes = 50
)

// App ...
type App struct {
	config     *configs.Config
	serverGRPC *grpc.Server
	repository *repository.Repo
	services   *service.Service
	controller *server.Server
}

// NewApp ...
func NewApp(ctx context.Context) (*App, error) {
	ctx = metadata.AppendToOutgoingContext(ctx, "service-name", ServiceName)

	c, err := configs.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("config.LoadConfig: %w", err)
	}

	app := &App{config: c}

	if err = app.initRepo(ctx); err != nil {
		return nil, err
	}
	app.initService()
	app.initServer()

	return app, nil
}

// ListenAndServeGRPC ...
func (app *App) ListenAndServeGRPC() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", app.config.Server.GRPCPort))
	if err != nil {
		return fmt.Errorf("net.Listen : %v", err)
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			middlewares.Validate,
		),
	)

	reflection.Register(grpcServer)
	pb.RegisterCommentsServer(grpcServer, app.controller)
	app.serverGRPC = grpcServer
	log.Printf("\napp bootstrap : %s \n", lis.Addr())

	return app.serverGRPC.Serve(lis)
}

// ListenAndServeHTTP ...
func (app *App) ListenAndServeHTTP() error {
	conn, err := grpc.NewClient(
		fmt.Sprintf(":%s", app.config.Server.GRPCPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("grpc.NewClient : %v", err)
	}

	gwmux := runtime.NewServeMux()
	ctx := context.Background()

	if err = pb.RegisterCommentsHandler(ctx, gwmux, conn); err != nil {
		return fmt.Errorf("RegisterCommentsHandler : %v", err)
	}

	gwServer := &http.Server{
		Addr:              fmt.Sprintf(":%s", app.config.Server.HTTPPort),
		Handler:           gwmux,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	log.Printf("http-gateway port : %s", gwServer.Addr)

	return gwServer.ListenAndServe()
}

// initRepo ...
func (app *App) initRepo(ctx context.Context) error {
	sm := shardmanager.NewShardManager(
		numNodes,
	)

	for _, cfg := range app.config.DBShards {
		dsn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
			cfg.User,
			cfg.Password,
			cfg.Host,
			cfg.Port,
			cfg.DBName,
		)

		db, err := pgxpool.New(ctx, dsn)
		if err != nil {
			return err
		}

		shardID, err := strconv.Atoi(cfg.ShardID)
		if err != nil {
			return err
		}

		sm.AddShard(db, shardID)
	}

	app.repository = repository.NewRepo(sm)

	return nil
}

// initService ...
func (app *App) initService() {
	app.services = service.NewService(app.repository)
}

// initServer ...
func (app *App) initServer() {
	app.controller = server.NewServer(app.services)
}

// Close ...
func (app *App) Close(ctx context.Context) {
	//nolint:errcheck, gosec
	_, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	app.services.Repository.Close()
	log.Println("connect repo closed")
	app.serverGRPC.GracefulStop()
	log.Println("GracefulStop success")
}
