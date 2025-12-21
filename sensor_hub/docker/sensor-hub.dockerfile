FROM golang:1.25-alpine

WORKDIR /app

COPY . ./
RUN go mod download

RUN go build -o sensor-hub .
RUN mv /app/sensor-hub /usr/local/bin/sensor-hub

RUN chmod +x ./docker/wait-for-mysql.sh

ENTRYPOINT ["/app/docker/wait-for-mysql.sh"]
CMD ["sensor-hub"]

