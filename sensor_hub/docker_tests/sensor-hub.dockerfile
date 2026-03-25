FROM golang:1.25-alpine

WORKDIR /app

COPY . ./
RUN go mod download

RUN go install github.com/go-delve/delve/cmd/dlv@latest

RUN go install github.com/air-verse/air@latest

CMD ["air", "-c", "docker_tests/air.toml"]