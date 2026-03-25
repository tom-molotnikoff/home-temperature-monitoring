# Stage 1: Build the React UI
FROM node:25-alpine AS ui-build
WORKDIR /ui
COPY ui/sensor_hub_ui/package.json ui/sensor_hub_ui/package-lock.json ./
RUN npm ci --silent
COPY ui/sensor_hub_ui/ ./
RUN npm run build

# Stage 2: Build the Go binary with embedded UI
FROM golang:1.25-alpine AS go-build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
COPY --from=ui-build /ui/dist ./web/dist/
RUN go build -o sensor-hub .

# Stage 3: Minimal runtime
FROM alpine:latest
WORKDIR /app
COPY --from=go-build /app/sensor-hub .
COPY --from=go-build /app/configuration/ ./configuration/
CMD ["./sensor-hub"]

