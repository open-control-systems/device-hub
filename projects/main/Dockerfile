FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY components ./components
COPY projects/main ./projects/main

WORKDIR /app/projects/main
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/device-hub .

FROM gcr.io/distroless/base-debian11
COPY --from=builder /app/device-hub /

CMD ["/device-hub"]
