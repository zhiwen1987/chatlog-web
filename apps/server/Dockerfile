# Chatlog Enterprise Server — Go 构建镜像
FROM golang:1.26-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/chatlog-server ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /out/chatlog-server /usr/local/bin/chatlog-server
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/chatlog-server"]