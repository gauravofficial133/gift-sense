FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata chromium nss freetype harfbuzz font-noto-emoji
WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /app/assets ./assets
ENV ROD_BROWSER_BIN=/usr/bin/chromium-browser
ENV CHROME_DEVEL_SANDBOX=/usr/local/sbin/chrome-devel-sandbox
EXPOSE 8080
CMD ["./server"]
