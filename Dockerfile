FROM golang:1.18
WORKDIR /build/anyapp
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o app cmd/main.go cmd/config.go cmd/helpers.go cmd/routes.go

FROM alpine
RUN apk --no-cache add curl net-tools
COPY --from=0 /build/anyapp/app .
EXPOSE 8080
ENTRYPOINT ["/app"]