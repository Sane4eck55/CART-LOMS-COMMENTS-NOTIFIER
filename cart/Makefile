BINDIR=${CURDIR}/bin
PACKAGE=route256/cart
LOMS_PROTO_DIR_PATH=../loms/api/v1
LOMS_PROTO_PATH=../loms/api/v1/loms.proto
LOMS_PB_PATH=github.com/Sane4eck55/CART-LOMS-COMMENTS-NOTIFIER/loms/api/v1


bindir:
	mkdir -p ${BINDIR}

build: bindir
	echo "build cart"

.PHONY: run-all
run-all:
	echo "starting build"
	docker-compose up --build

.PHONY: test
test:
	go test -timeout 30s -race ./... -count=1

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
	rm -rf ./tests/allure-results
	go test -tags=e2e ./tests/... -count=1 

.PHONY: allure
allure: 
	allure serve ./tests/allure-results

.PHONY: lint-complexity
lint-complexity:
#если папки нет, то создаем ее
	@if [ ! -d "./tests/reports/linter" ]; then \
		mkdir -p ./tests/reports/linter; \
	fi
	rm -fr ./tests/reports/linter/*
#tee нужно для созранения вывода команды слева от | (например find . -name '*.go' -print0 | xargs -0 gocyclo -over 10)
#-print0 заставляет find выводить имена файлов, разделяя их нулевым байтом (\0), а не \n
#xargs читает входные данные и формирует команды, подставляя аргументы.
#-0 говорит xargs использовать нулевой байт (\0) как разделитель аргументов, а не пробелы или переводы строк.
	find . ! -name '*_mock.go' -name '*.go' -print0 | xargs -0 gocyclo -over 10 | tee ./tests/reports/linter/gocyclo-reports.txt
	find . ! -name '*_mock.go' -name '*.go' -print0 | xargs -0 gocognit -over 15 | tee ./tests/reports/linter/gocognit-reports.txt
#если файлы не пустые значит есть функции с высокой сложностью, команда test ! -s вернёт ошибку, и CI упадёт.
	test ! -s ./tests/reports/linter/gocyclo-reports.txt
	test ! -s ./tests/reports/linter/gocognit-reports.txt

.PHONY: generate-mocks
generate-mocks: generate-mocks-server generate-mocks-service

.PHONY: generate-mocks-server
generate-mocks-server:
	rm -fr ./internal/app/server/mocks/*.go
	minimock -i ./internal/app/server.* -o ./internal/app/server/mocks/ -s "_mock.go"

.PHONY: generate-mocks-service
generate-mocks-service:
	rm -fr ./internal/domain/service/mocks/*.go
	minimock -i ./internal/domain/service.* -o ./internal/domain/service/mocks/ -s "_mock.go"

.PHONY: bench
bench:
	go test -bench=BenchmarkInMemoryRepository_Add ./internal/domain/repository -benchmem -v


.PHONY: bin-deps
bin-deps:
	GOBIN=$(BINDIR) go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.1 && \
	GOBIN=$(BINDIR) go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2.0 && \
	GOBIN=$(BINDIR) go install github.com/envoyproxy/protoc-gen-validate@v1.0.4 

.PHONY: vendor-rm
vendor-rm:
	rm -rf vendor-proto

.PHONY: vendor-proto
vendor-proto: vendor-rm vendor-proto/protoc-gen-openapiv2/options vendor-proto/google/protobuf vendor-proto/google/api vendor-proto/validate

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
protoc-generate: bin-deps vendor-proto create-pkg-dir
	protoc \
	-I ${LOMS_PROTO_DIR_PATH} \
	-I vendor-proto \
	--plugin=protoc-gen-go=$(BINDIR)/protoc-gen-go \
	--go_out pkg/${LOMS_PB_PATH} \
	--go_opt paths=source_relative \
	--plugin=protoc-gen-go-grpc=$(BINDIR)/protoc-gen-go-grpc \
	--go-grpc_out pkg/${LOMS_PB_PATH} \
	--go-grpc_opt paths=source_relative \
	--plugin=protoc-gen-validate=$(BINDIR)/protoc-gen-validate \
	--validate_out="lang=go,paths=source_relative:pkg/${LOMS_PB_PATH}" \
	${LOMS_PROTO_PATH}
	go mod tidy

.PHONY: create-pkg-dir
create-pkg-dir:
	mkdir -p pkg/${LOMS_PB_PATH}
