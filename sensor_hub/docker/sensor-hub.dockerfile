FROM golang:1.25-alpine

WORKDIR /app

COPY ./go.mod ./go.sum ./
RUN go mod download

COPY . .

RUN rm ./database.properties

COPY docker/write-db-properties-and-wait.sh /app/
RUN chmod +x /app/write-db-properties-and-wait.sh

RUN go build -o sensor-hub .

ENTRYPOINT ["/app/write-db-properties-and-wait.sh"]
CMD ["./sensor-hub"]