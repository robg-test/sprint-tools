# Stage 1: Build
FROM golang:1.24 AS builder

RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - && apt-get install -y nodejs

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY package.json package-lock.json ./
RUN npm install

COPY . .

RUN go tool templ generate && \
    npx @tailwindcss/cli -i ./web/static/css/input.css -o ./web/static/css/output.css && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o sprint-tools

# Stage 2: Runtime
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/sprint-tools /app/sprint-tools
COPY --from=builder /app/web/static /app/web/static

ENV ENV=production

EXPOSE 443

CMD ["./sprint-tools"]
