FROM golang:1.25-alpine

WORKDIR /app

COPY . ./
RUN go mod download

RUN go install github.com/go-delve/delve/cmd/dlv@latest

RUN go install github.com/air-verse/air@latest

RUN chmod +x ./docker/wait-for-mysql.sh

ENTRYPOINT ["/app/docker/wait-for-mysql.sh"]
CMD ["./sensor-hub"]