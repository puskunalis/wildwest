FROM golang:1.20.2-alpine3.17 AS build

WORKDIR /app

COPY . .

RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -w" -o bin/cowboy cmd/cowboy/main.go

FROM gcr.io/distroless/static:nonroot

WORKDIR /app

EXPOSE 8080
EXPOSE 50051

COPY --from=build /app/bin/cowboy .

CMD ["/app/cowboy"]
