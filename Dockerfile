FROM golang AS builder

COPY . /app

WORKDIR /app

RUN go mod vendor
RUN CGO_ENABLED=0 go build -ldflags="-s -w -X main.build=$(git describe --tags --exact-match HEAD 2>/dev/null || git rev-parse --short HEAD 2>/dev/null || echo -n unknown)" -o /exporter-test ./main.go


FROM scratch

COPY --from=builder /exporter-test /

EXPOSE 8080

ENTRYPOINT ["/extporter-test"]
