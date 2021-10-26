FROM golang:1.17.2-stretch AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

WORKDIR /app/backend/cmd
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s" -o /go/bin/backend


FROM debian:stretch-20211011-slim

COPY --from=builder /go/bin/backend /

CMD ["/backend"]
