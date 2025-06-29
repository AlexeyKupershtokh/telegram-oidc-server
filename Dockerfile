FROM golang:1.24.3 AS builder

WORKDIR /usr/src/app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o /usr/local/bin/app ./cmd/service

FROM gcr.io/distroless/static-debian12

COPY --from=builder /usr/local/bin/app /usr/local/bin/app

CMD ["app"]
