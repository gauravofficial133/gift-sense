FROM golang:1.23-alpine@sha256:383395b794dffa5b53012a212365d40c8e37109a626ca30d6151c8348d380b5f AS builder
WORKDIR /app
RUN apk add --no-cache gcc musl-dev
COPY go.mod go.sum ./
RUN go mod download
RUN go mod download github.com/mattn/go-sqlite3
COPY . .
RUN GONOSUMDB=* CGO_ENABLED=1 GOOS=linux go build -mod=mod -o server ./cmd/server

FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata chromium nss freetype harfbuzz font-noto-emoji
WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /app/assets ./assets
ENV ROD_BROWSER_BIN=/usr/bin/chromium-browser
ENV CHROME_DEVEL_SANDBOX=/usr/local/sbin/chrome-devel-sandbox
EXPOSE 8080
CMD ["./server"]
