FROM golang:1.23-alpine AS builder

WORKDIR /app
RUN apk --no-cache add bash git make

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o ./bin/app cmd/main.go

FROM alpine AS runner

COPY --from=builder /app/bin/app /app
CMD ["/app"]
