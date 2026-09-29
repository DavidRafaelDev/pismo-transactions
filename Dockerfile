# --- builder ---
FROM golang:1.25-alpine AS builder
WORKDIR /src

# cache dependencies layer (go.sum may not exist yet in early stages)
COPY go.mod ./
COPY go.sum* ./
RUN go mod download

# compile static binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

# --- runtime ---
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=builder /out/api /app/api
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/api"]
