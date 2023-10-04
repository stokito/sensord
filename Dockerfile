FROM golang:1.21
WORKDIR /src
RUN go build -o /build/sensord ./cmd/sensord/main.go

FROM alpine
WORKDIR /opt/
COPY /build/sensord /opt/
EXPOSE 8080
EXPOSE 9090
CMD ["/opt/sensord"]