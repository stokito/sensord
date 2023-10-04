FROM golang:1.21 AS builder
WORKDIR /src/
COPY ./ ./
RUN go build -o ./build/sensord ./cmd/sensord/main.go

FROM alpine
WORKDIR /opt/
COPY --from=builder /src/build/sensord /opt/sensord
EXPOSE 8080
EXPOSE 9090
CMD ["/opt/sensord"]