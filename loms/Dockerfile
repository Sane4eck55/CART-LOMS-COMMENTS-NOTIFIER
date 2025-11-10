FROM golang:1.23.4-alpine AS builder

WORKDIR /build

COPY ./cart ./cart
COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server/main.go

FROM scratch
COPY --from=builder server /bin/server
#COPY configs/values_local.yaml /bin/config/values_local_docker.yaml
COPY configs/values_ci.yaml /bin/loms/configs/values_ci.yaml

#ENV CONFIG_FILE=/bin/config/values_local_docker.yaml
ENV CONFIG_FILE=/bin/loms/configs/values_ci.yaml

ENTRYPOINT ["/bin/server"]