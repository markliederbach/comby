FROM golang:1.22-alpine as builder
ARG VERSION=latest

RUN adduser --uid 1500 -D comby

WORKDIR /app
COPY ./go.mod ./go.mod
COPY ./go.sum ./go.sum
COPY ./client ./client
COPY ./command ./command
COPY ./util ./util
COPY ./main.go ./main.go

ENV CGO_ENABLED=0
RUN go mod download && \
    go build -ldflags "-X main.version=$VERSION" -o /app/comby

FROM scratch
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/comby /comby

ENTRYPOINT ["/comby"]