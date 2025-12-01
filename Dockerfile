FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod ./
RUN go mod download

# Copy source
COPY . ./

# Build (go mod tidy will create go.sum if needed)
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -o /videogames2

# Runtime
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /videogames2 .
COPY static/ ./static/

EXPOSE 8080

CMD ["./videogames2"]
