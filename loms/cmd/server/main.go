// Package main ...
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/pkg/logger"
	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/cart/pkg/tracer"
	config "github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/loms/configs"
	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/loms/internal/app/server"
	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/loms/internal/outbox"
	"github.com/go-chi/cors"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	mw "github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/loms/internal/middlewares"
	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/loms/internal/service"
	pb "github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/loms/pkg/api/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"

	"github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/loms/internal/repository/postgres/connect"
	repo "github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/loms/internal/repository/sqlc"

	"github.com/IBM/sarama"

	serviceproducer "github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/loms/internal/kafka/producer"
)

const (
	// readHeaderTimeout ...
	readHeaderTimeout = 1 * time.Minute
	// ServiceName ...
	ServiceName = "loms"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger.Infow("started server loms")
	ctx = metadata.AppendToOutgoingContext(ctx, "service-name", ServiceName)

	t, err := tracer.NewTracer(ctx)
	if err != nil {
		logger.Fatalw(fmt.Sprintf("NewTracer : %v", err))
	}

	defer func() {
		//nolint:govet
		if err := t.TracerProvider.Shutdown(ctx); err != nil {
			logger.Fatalw(fmt.Sprintf("TracerProvider.Shutdown : %v", err))
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	logger.GetLogger(ServiceName)
	defer func() {
		_ = logger.Sync()
	}()

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatalw(fmt.Sprintf("LoadConfig : %v", err))
	}
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.Server.GRPCPort))
	if err != nil {
		logger.Fatalw(fmt.Sprintf("Listen : %v", err))
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			mw.Validate,
			mw.UnaryInterceptor,
		),
	)

	reflection.Register(grpcServer)

	masterPool, replicaPool, err := createPool(cfg)
	if err != nil {
		logger.Fatalw(fmt.Sprintf("createPool : %v", err))
	}
	defer func() {
		masterPool.Close()
		replicaPool.Close()
	}()

	repo, err := initDB(masterPool, replicaPool, t.Tracer)
	if err != nil {
		logger.Fatalw(fmt.Sprintf("initDB : %v", err))
	}

	producer, err := initKafkaProducer(cfg.Kafka.Brokers)
	if err != nil {
		logger.Fatalw(fmt.Sprintf("initKafka : %v", err))
	}

	defer func() {
		if err = producer.Close(); err != nil {
			logger.Errorw(fmt.Sprintf("Ошибка при закрытии продюсера: %v", err))
		}
		log.Println("producer.Close success")
	}()

	producerOrderEvent := serviceproducer.NewProducer(producer, cfg.Kafka.TopicName)

	services := service.NewService(repo, t.Tracer, producerOrderEvent)
	controller := server.NewServer(&services, t.Tracer)

	outbox := outbox.NewOutbox(ctx)
	outbox.Start(&services)

	pb.RegisterLomsServer(grpcServer, controller)

	logger.Infow(fmt.Sprintf("server listening at %v", lis.Addr()))

	go func() {
		// nolint:govet
		conn, err := grpc.NewClient(
			fmt.Sprintf(":%s", cfg.Server.GRPCPort),
			grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			logger.Fatalw(fmt.Sprintf("grpc.NewClient : %v", err))
		}

		gwmux := runtime.NewServeMux()
		if err = pb.RegisterLomsHandler(context.Background(), gwmux, conn); err != nil {
			log.Fatalln("RegisterLomsHandler :", err)
		}

		c := initCors()

		mux := http.NewServeMux()
		mux.Handle("/", gwmux)
		mux.Handle("/metrics", promhttp.Handler())

		gwServer := &http.Server{
			Addr:              fmt.Sprintf(":%s", cfg.Server.HTTPPort),
			Handler:           c.Handler(mux),
			ReadHeaderTimeout: readHeaderTimeout,
		}

		logger.Infow(fmt.Sprintf("Serving gRPC-Gateway on %s\n", gwServer.Addr))
		if err = gwServer.ListenAndServe(); err != nil {
			logger.Fatalw(fmt.Sprintf("gwServer.ListenAndServe : %v", err))
		}
	}()

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			logger.Fatalw(fmt.Sprintf("grpcServer.Serve: %v", err))
		}
	}()

	<-stop
	logger.Infow("Shutting down gracefully...")
	outbox.Stop()
	logger.Infow("outbox.Stop success")

	grpcServer.GracefulStop()
	logger.Infow("grpcServer.GracefulStop success")

	cancel()
	logger.Infow("ctx cancel")
}

// initCors ...
func initCors() *cors.Cors {
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:8089"}, // или "*" для всех
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})
	return c
}

// initDB ...
func initDB(poolMaster *pgxpool.Pool, poolReplica *pgxpool.Pool, tracer service.Tracer) (*repo.Repo, error) {
	repo := repo.NewRepo(poolMaster, poolReplica, tracer)

	return repo, nil
}

// createPool ...
func createPool(cfg *config.Config) (*pgxpool.Pool, *pgxpool.Pool, error) {
	dsnMaster := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DataBaseMaster.User,
		cfg.DataBaseMaster.Password,
		cfg.DataBaseMaster.Host,
		cfg.DataBaseMaster.Port,
		cfg.DataBaseMaster.DBName,
	)

	poolMaster, err := connect.NewPool(context.TODO(), dsnMaster)
	if err != nil {
		return nil, nil, fmt.Errorf("NewPool: %w", err)
	}

	dsnReplica := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DataBaseReplica.User,
		cfg.DataBaseReplica.Password,
		cfg.DataBaseReplica.Host,
		cfg.DataBaseReplica.Port,
		cfg.DataBaseReplica.DBName,
	)

	poolReplica, err := connect.NewPool(context.TODO(), dsnReplica)
	if err != nil {
		return nil, nil, fmt.Errorf("NewPool: %w", err)
	}

	return poolMaster, poolReplica, nil
}

// initKafkaProducer ...
func initKafkaProducer(broker string) (sarama.SyncProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Producer.Partitioner = sarama.NewHashPartitioner
	config.Producer.RequiredAcks = sarama.WaitForAll
	producer, err := sarama.NewSyncProducer([]string{broker}, config)
	if err != nil {
		return nil, err
	}

	return producer, nil
}
