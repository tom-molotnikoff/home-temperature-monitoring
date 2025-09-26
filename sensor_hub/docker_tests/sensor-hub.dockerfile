FROM golang:1.25-alpine

WORKDIR /app

COPY ./go.mod ./go.sum ./
RUN go mod download

RUN go install github.com/go-delve/delve/cmd/dlv@latest

RUN go install github.com/air-verse/air@latest

COPY . ./

RUN chmod +x ./docker_tests/wait-for-mysql.sh

ENTRYPOINT ["/app/docker_tests/wait-for-mysql.sh"]
CMD ["air", "-c", "docker_tests/air.toml"]