FROM golang:1.20.2-alpine3.17 AS build

WORKDIR /app

COPY . .

RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -w" -o bin/cowboy-controller cmd/cowboy-controller/main.go

FROM gcr.io/distroless/static:nonroot

WORKDIR /app

EXPOSE 8080

COPY --from=build /app/bin/cowboy-controller .

CMD ["/app/cowboy-controller"]
