BINDIR=${CURDIR}/bin
PACKAGE=route256/comments

COMMENTS_PROTO_PATH := "api/v1"
ifdef LOCAL
	MIGRATIONS_FOLDER := $(CURDIR)/migrations/
	DB_DSN_SHARD_1 := "postgresql://comments-user-1:comments-password-1@localhost:5434/comments_db?sslmode=disable"
	DB_DSN_SHARD_2 := "postgresql://comments-user-2:comments-password-2@localhost:5435/comments_db?sslmode=disable"
else
	MIGRATIONS_FOLDER := $(CURDIR)/comments/migrations/
	DB_DSN_SHARD_1 := "postgresql://comments-user-1:comments-password-1@postgres-comments-shard-1:5432/comments_db?sslmode=disable"
	DB_DSN_SHARD_2 := "postgresql://comments-user-2:comments-password-2@postgres-comments-shard-2:5432/comments_db?sslmode=disable"
endif

bindir:
	mkdir -p ${BINDIR}

build: bindir
	echo "build comments"

.PHONY: run-migrations
run-migrations: goose-migrate-up-all
	echo "migrations finished"

.PHONY: run
run:
	echo "starting build"
	docker-compose up --build

.PHONY: bin-deps
bin-deps: install-goose
	GOBIN=$(BINDIR) go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.1 && \
	GOBIN=$(BINDIR) go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2.0 && \
	GOBIN=$(BINDIR) go install github.com/envoyproxy/protoc-gen-validate@v1.0.4 && \
	GOBIN=$(BINDIR) go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.19.1 && \
	GOBIN=$(BINDIR) go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.19.1 && \
	GOBIN=$(BINDIR) go install github.com/go-swagger/go-swagger/cmd/swagger@v0.30.5 

.PHONY: vendor-rm
vendor-rm:
	rm -rf vendor-proto

.PHONY: vendor-proto
vendor-proto: vendor-rm vendor-proto/google/protobuf vendor-proto/google/api vendor-proto/validate vendor-proto/protoc-gen-openapiv2/options

# Устанавливаем proto описания google/protobuf
.PHONY: vendor-proto/google/protobuf
vendor-proto/google/protobuf:
	git clone -b main --single-branch -n --depth=1 --filter=tree:0 \
		https://github.com/protocolbuffers/protobuf vendor-proto/protobuf &&\
	cd vendor-proto/protobuf &&\
	git sparse-checkout set --no-cone src/google/protobuf &&\
	git checkout
	mkdir -p vendor-proto/google
	mv vendor-proto/protobuf/src/google/protobuf vendor-proto/google
	rm -rf vendor-proto/protobuf

# Устанавливаем proto описания validate
.PHONY: vendor-proto/validate
vendor-proto/validate:
	git clone -b main --single-branch --depth=2 --filter=tree:0 \
		https://github.com/bufbuild/protoc-gen-validate vendor-proto/tmp && \
		cd vendor-proto/tmp && \
		git sparse-checkout set --no-cone validate &&\
		git checkout
		mkdir -p vendor-proto/validate
		mv vendor-proto/tmp/validate vendor-proto/
		rm -rf vendor-proto/tmp

# Устанавливаем proto описания google/googleapis
.PHONY: vendor-proto/google/api
vendor-proto/google/api:
	git clone -b master --single-branch -n --depth=1 --filter=tree:0 \
 		https://github.com/googleapis/googleapis vendor-proto/googleapis && \
 	cd vendor-proto/googleapis && \
	git sparse-checkout set --no-cone google/api && \
	git checkout
	mkdir -p  vendor-proto/google
	mv vendor-proto/googleapis/google/api vendor-proto/google
	rm -rf vendor-proto/googleapis	

# Устанавливаем proto описания protoc-gen-openapiv2/options
.PHONY: vendor-proto/protoc-gen-openapiv2/options
vendor-proto/protoc-gen-openapiv2/options:
	git clone -b main --single-branch -n --depth=1 --filter=tree:0 \
 		https://github.com/grpc-ecosystem/grpc-gateway vendor-proto/grpc-ecosystem && \
 	cd vendor-proto/grpc-ecosystem && \
	git sparse-checkout set --no-cone protoc-gen-openapiv2/options && \
	git checkout
	mkdir -p vendor-proto/protoc-gen-openapiv2
	mv vendor-proto/grpc-ecosystem/protoc-gen-openapiv2/options vendor-proto/protoc-gen-openapiv2
	rm -rf vendor-proto/grpc-ecosystem

.PHONY: protoc-generate
protoc-generate: bin-deps vendor-proto
	protoc \
		-I ${COMMENTS_PROTO_PATH} \
		-I vendor-proto \
		--plugin=protoc-gen-go=$(BINDIR)/protoc-gen-go \
		--go_out pkg/${COMMENTS_PROTO_PATH} \
		--go_opt paths=source_relative \
		--plugin=protoc-gen-go-grpc=$(BINDIR)/protoc-gen-go-grpc \
		--go-grpc_out pkg/${COMMENTS_PROTO_PATH} \
		--go-grpc_opt paths=source_relative \
		--plugin=protoc-gen-validate=$(BINDIR)/protoc-gen-validate \
		--validate_out="lang=go,paths=source_relative:pkg/api/v1" \
		--plugin=protoc-gen-grpc-gateway=$(BINDIR)/protoc-gen-grpc-gateway \
		--grpc-gateway_out pkg/${COMMENTS_PROTO_PATH} \
		--grpc-gateway_opt logtostderr=true --grpc-gateway_opt paths=source_relative --grpc-gateway_opt generate_unbound_methods=true \
		--plugin=protoc-gen-openapiv2=$(BINDIR)/protoc-gen-openapiv2 \
		--openapiv2_out pkg/api/v1 \
		--openapiv2_opt logtostderr=true \
		api/v1/comments.proto

.PHONY: install-goose
install-goose:
	$(info Installing goose binary into [$(BINDIR)]...)
	GOBIN=$(BINDIR) go install github.com/pressly/goose/v3/cmd/goose@v3.24.1

.PHONY: goose-migrate-up-all
goose-migrate-up-all: install-goose goose-migrate-up-1 goose-migrate-up-2

.PHONY: goose-migrate-down-all
goose-migrate-down-all: goose-migrate-down-1 goose-migrate-down-2

.PHONY: goose-migrate-up-1
goose-migrate-up-1:
	$(BINDIR)/goose -dir $(MIGRATIONS_FOLDER) postgres "$(DB_DSN_SHARD_1)" up

.PHONY: goose-migrate-up-2
goose-migrate-up-2:
	$(BINDIR)/goose -dir $(MIGRATIONS_FOLDER) postgres "$(DB_DSN_SHARD_2)" up

.PHONY: goose-migrate-down-1
goose-migrate-down-1:
	$(BINDIR)/goose -dir $(MIGRATIONS_FOLDER) postgres "$(DB_DSN_SHARD_1)" down

.PHONY: goose-migrate-down-2
goose-migrate-down-2:
	$(BINDIR)/goose -dir $(MIGRATIONS_FOLDER) postgres "$(DB_DSN_SHARD_2)" down

.PHONY: generate-mocks
generate-mocks: generate-mocks-server generate-mocks-service

.PHONY: generate-mocks-server
generate-mocks-server:
	rm -fr ./internal/app/server/mocks/*.go
	minimock -i ./internal/app/server.* -o ./internal/app/server/mocks/ -s "_mock.go"

.PHONY: generate-mocks-service
generate-mocks-service:
	rm -fr ./internal/service/mocks/*.go
	minimock -i ./internal/service.* -o ./internal/service/mocks/ -s "_mock.go"

.PHONY: test
test:
	go test ./... -count=1