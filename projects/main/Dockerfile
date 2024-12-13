FROM golang:1.23-bookworm AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY components ./components
COPY projects/main ./projects/main

WORKDIR /app/projects/main

RUN go build -o /app/device-hub .

FROM gcr.io/distroless/base-debian12
COPY --from=builder /app/device-hub /

CMD ["/device-hub"]
