FROM golang:1.26.5 AS builder

WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 go build -o /bin/main cmd/main.go

FROM scratch
COPY --from=builder /bin/main /bin/main

CMD ["/bin/main"]
