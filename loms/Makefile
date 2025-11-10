BINDIR=${CURDIR}/bin
PACKAGE=route256/loms

LOMS_PROTO_PATH := "api/v1"
KAFKA_PROTO_PATH := "api/kafka"

ifdef LOCAL
	LOCAL_DB_DSN := postgresql://user:password@127.0.0.1:5432/route256?sslmode=disable
	MIGRATIONS_FOLDER := $(CURDIR)/migrations/
	TEST_MIGRATIONS_FOLDER := $(CURDIR)/tests/repository/migrations/
else
	LOCAL_DB_DSN := postgresql://loms-user:loms-password@postgres-master:5432/loms_db?sslmode=disable
	MIGRATIONS_FOLDER := $(CURDIR)/loms/migrations/
endif

bindir:
	mkdir -p ${BINDIR}

build: bindir
	echo "build loms"

run-migrations: install-goose db-migrate-up
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
		-I ${LOMS_PROTO_PATH} \
		-I vendor-proto \
		--plugin=protoc-gen-go=$(BINDIR)/protoc-gen-go \
		--go_out pkg/${LOMS_PROTO_PATH} \
		--go_opt paths=source_relative \
		--plugin=protoc-gen-go-grpc=$(BINDIR)/protoc-gen-go-grpc \
		--go-grpc_out pkg/${LOMS_PROTO_PATH} \
		--go-grpc_opt paths=source_relative \
		--plugin=protoc-gen-validate=$(BINDIR)/protoc-gen-validate \
		--validate_out="lang=go,paths=source_relative:pkg/api/v1" \
		--plugin=protoc-gen-grpc-gateway=$(BINDIR)/protoc-gen-grpc-gateway \
		--grpc-gateway_out pkg/${LOMS_PROTO_PATH} \
		--grpc-gateway_opt logtostderr=true --grpc-gateway_opt paths=source_relative --grpc-gateway_opt generate_unbound_methods=true \
		--plugin=protoc-gen-openapiv2=$(BINDIR)/protoc-gen-openapiv2 \
		--openapiv2_out pkg/api/v1 \
		--openapiv2_opt logtostderr=true \
		api/v1/loms.proto

	protoc \
		-I ${KAFKA_PROTO_PATH} \
		-I vendor-proto \
		--plugin=protoc-gen-go=$(BINDIR)/protoc-gen-go \
		--go_out=pkg/${KAFKA_PROTO_PATH} \
		--go_opt=paths=source_relative \
		--plugin=protoc-gen-go-grpc=$(BINDIR)/protoc-gen-go-grpc \
		--go-grpc_out=pkg/${KAFKA_PROTO_PATH} \
		--go-grpc_opt=paths=source_relative \
		--plugin=protoc-gen-validate=$(BINDIR)/protoc-gen-validate \
		--validate_out=lang=go,paths=source_relative:pkg/${KAFKA_PROTO_PATH} \
		--plugin=protoc-gen-grpc-gateway=$(BINDIR)/protoc-gen-grpc-gateway \
		--grpc-gateway_out=pkg/${KAFKA_PROTO_PATH} \
		--grpc-gateway_opt logtostderr=true --grpc-gateway_opt=paths=source_relative --grpc-gateway_opt=generate_unbound_methods=true \
		api/kafka/order-events.proto

	go mod tidy

.PHONY: swagger
swagger:
	bin/swagger serve ./api/openapiv2/loms.swagger.json

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

.PHONY: test-coverage
test-coverage:
#если папки нет, то создаем ее
	@if [ ! -d "./tests/reports/coverage" ]; then \
		mkdir -p ./tests/reports/coverage; \
	fi
	rm -fr ./tests/reports/coverage/*
	go test -coverprofile=./tests/reports/coverage/coverage.out ./...
	go tool cover -html=./tests/reports/coverage/coverage.out -o ./tests/reports/coverage/coverage.html

.PHONY: api-tests
api-tests:
	rm -rf ./tests/server/allure-results
	rm -rf ./tests/repository/allure-results
	$(MAKE) test-db-migrate-up
	@set -e; \
	result=0; \
	go test -tags=e2e ./tests/... -count=1 || result=$$?; \
	$(MAKE) test-db-migrate-down; \
	$(MAKE) db-migrate-down; \
	$(MAKE) db-migrate-up; \
	exit $$result

.PHONY: install-goose
install-goose:
	$(info Installing goose binary into [$(BINDIR)]...)
	GOBIN=$(BINDIR) go install github.com/pressly/goose/v3/cmd/goose@v3.24.1

.PHONY: db-create-migration
db-create-migration:
	$(BINDIR)/goose -dir $(MIGRATIONS_FOLDER) create -s $(n) sql

.PHONY: db-migrate-up
db-migrate-up:
	$(BINDIR)/goose -dir $(MIGRATIONS_FOLDER) postgres "$(LOCAL_DB_DSN)" up

.PHONY: db-migrate-down
db-migrate-down:
	$(BINDIR)/goose -dir $(MIGRATIONS_FOLDER) postgres "$(LOCAL_DB_DSN)" down

.PHONY: test-db-migrate-up
test-db-migrate-up:
	$(BINDIR)/goose -dir $(TEST_MIGRATIONS_FOLDER) postgres "$(LOCAL_DB_DSN)" up

.PHONY: test-db-migrate-down
test-db-migrate-down:
	$(BINDIR)/goose -dir $(TEST_MIGRATIONS_FOLDER) postgres "$(LOCAL_DB_DSN)" down