# Описание protoc-generate в Makefile

```
.PHONY: bin-deps
bin-deps:
#устанаваливаем генератор гошных структур по message
	GOBIN=$(LOCAL_BIN) go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.1 && \
#устанаваливаем генератор Go кода для grpc	
	GOBIN=$(LOCAL_BIN) go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2.0 &&
```

```
.PHONY: protoc-generate
protoc-generate:
#запускаем protoc
	protoc \  
#путь до папки с прото файлами
	-I ${LOMS_PROTO_PATH} \
	-I vendor-proto \

#запускаем плагин для генерации гошных структуру на основе message
	--plugin=protoc-gen-go=$(LOCAL_BIN)/protoc-gen-go \

#путь куда будет генерить protoc-gen-go	
	--go_out pkg/${LOMS_PROTO_PATH} \

#нужно чтобы сохранить структуру папок api/v1 в папке pkg при генерации protoc-gen-go	
	--go_opt paths=source_relative \

#	
	--plugin=protoc-gen-go-grpc=$(LOCAL_BIN)/protoc-gen-go-grpc \
	--go-grpc_out pkg/${LOMS_PROTO_PATH} \
	--go-grpc_opt paths=source_relative \
	--plugin=protoc-gen-validate=$(LOCAL_BIN)/protoc-gen-validate \
    --validate_out="lang=go,paths=source_relative:pkg/api/notes/v1" \
	--plugin=protoc-gen-grpc-gateway=$(LOCAL_BIN)/protoc-gen-grpc-gateway \
    --grpc-gateway_out pkg/${LOMS_PROTO_PATH} \
    --grpc-gateway_opt logtostderr=true  \
	--grpc-gateway_opt paths=source_relative \
	--grpc-gateway_opt generate_unbound_methods=true \
	--plugin=protoc-gen-openapiv2=$(LOCAL_BIN)/protoc-gen-openapiv2 \
    --openapiv2_out api/openapiv2 \
    --openapiv2_opt logtostderr=true \
#путь до прото файла
	api/v1/loms.proto && \
	go mod tidy

```

# api-tests
```
.PHONY: api-tests
api-tests:
	rm -rf ./tests/server/allure-results
	rm -rf ./tests/repository/allure-results
	$(MAKE) test-db-migrate-up
	@set -e; \
	result=0; \
	go test -tags=e2e ./tests/... -count=1 || result=$$?; \
	$(MAKE) test-db-migrate-down; \
	exit $$result


@set -e; \ — чтобы shell прервался при ошибках;
result=$$?; \ — сохраняем статус go test;
exit $$result — передаем код возврата из go test.
```

# Local env
```
LOCAL=1 make run-migrations
при локально запуску нужно указывать LOCAL=1, чтоб подзватывался локальный DSN 
```



# Для взаимодейсвтия двух контейнеров
поднимаем сеть
```  
docker network create shared_net
```

### Cart docker-compose.yaml
``` 
version: "3.7"

services:
  cart:
    build: .
    ports:
      - "8080:8080"
    depends_on:
      - products
    networks:
      - shared_net
  products:
    image: gitlab-registry.ozon.dev/go/classroom-20/students/homework-draft/products:latest
    ports:
      - "8082:8082"
    networks:
      - shared_net
networks:
  shared_net:
    external: true
``` 

### Loms docker-compose.yaml
```
version: "3.7"

services:
  loms:
    build: .
    ports:
      - "8084:8084"
      - "50051:50051" 
    networks:
      - shared_net

  swagger-ui:
    image: swaggerapi/swagger-ui
    ports:
      - "8085:8085"
    environment:
      - SWAGGER_JSON=/swagger/loms.swagger.json
    volumes:
      - ./pkg/api/v1/loms.swagger.json:/swagger/loms.swagger.json:ro
    networks:
      - shared_net
   
networks:
  shared_net:
    external: true
```


# CORS
```
Когда вы открываете сайт (например, example_1.ru), браузер по умолчанию не разрешает этому сайту делать запросы к другим доменам (например, sports.ru) — это ограничение для защиты от вредоносных действий.

CORS — это набор правил, по которым сервер сообщает браузеру, что его ресурсы разрешены для вызова со сторонних сайтов.
```
### логика работы
```
1. Браузер посылает "пред-запрос" (preflight request) – это OPTIONS-запрос с информацией о запросе, чтобы проверить, разрешено ли выполнение реального запроса.
2. Сервер отвечает с определёнными заголовками, например:
Access-Control-Allow-Origin: http://example_1.ru
Access-Control-Allow-Methods: GET, POST
3. Если сервер разрешил, браузер выполняет оригинальный запрос.
4. Если нет — запрос блокируется, и пользователь видит ошибку.
```

### Важные заголовки CORS
```
Access-Control-Allow-Origin — указывает, какие источники могут делать запросы (например, http://localhost:8089 или * — все источники).
Access-Control-Allow-Methods — какие HTTP методы разрешены (GET, POST, PUT, DELETE).
Access-Control-Allow-Headers — какие заголовки разрешены при запросе.
Access-Control-Allow-Credentials — разрешить ли отправку cookies или авторизационных данных.
```

# Swagger
```
  "host": "localhost:8084",
  "basePath": "/",
  "schemes": [
    "http"
  ],
```

# Postgesql

```
psql 'postgresql://user:password@localhost:5433/route256?sslmode=disable'
```

# Goose
```
./bin/goose -dir migrations create stocks_table sql
// 20250916133317_stocks_table.sql
```

```
./bin/goose -s -dir ./tests/repository/migrations create order_table sql
./bin/goose -s -dir migrations create order_table sql
// 00001_order_table.sql , следующая будет 0002_
```

```
.PHONY: db-create-migration
db-create-migration:
	$(BINDIR)/goose -dir $(MIGRATIONS_FOLDER) create -s $(n) sql

export n='new_migrations' && make db-create-migration
 ``` 

```
create table orders_items(
  order_id bigserial not null REFERENCES orders(id) ON DELETE CASCADE,
);

REFERENCES orders(id) - связывает поле order_id с полем id из таблички orders
ON DELETE CASCADE - если удалить строчку с id из таблицы orders, то из таблицы orders_items будут удалены  строчки где order_id = id 

лучше не использовать REFERENCES orders(id) ON DELETE CASCADE, что в дальнейшен было легче шардировать 
```


# FOR UPDATE
```
При выполнении запроса  с FOR UPDATE, база данных заблокирует строки, которые выбраны этим запросом.
Эти строки не смогут быть изменены (обновлены или удалены) другими транзакциями, пока ваша транзакция не завершится (COMMIT или ROLLBACK).
Другие транзакции, пытающиеся изменить эти же строки, будут ждать освобождения блокировки или получат ошибку в зависимости от уровня изоляции и настроек.
```