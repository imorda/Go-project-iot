# syntax=docker/dockerfile:1

FROM golang:1.22 as builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -o home-controller cmd/server/main.go

RUN chmod +x ./home-controller


FROM alpine:edge

WORKDIR /app

RUN apk --no-cache add openssl && apk --no-cache add ca-certificates

COPY --from=builder /app/home-controller .

EXPOSE 8080
EXPOSE 80

CMD ["./home-controller"]
