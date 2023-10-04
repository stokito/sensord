FROM golang:1.21 AS builder
WORKDIR /src/
COPY ./ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o ./build/sensord ./cmd/sensord/main.go

FROM alpine
WORKDIR /opt/
COPY --from=builder /src/build/sensord /opt/sensord
EXPOSE 8080
EXPOSE 9090
CMD ["/opt/sensord"]