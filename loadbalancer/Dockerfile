
FROM golang:1.23.5-alpine AS build

WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o /goapp



FROM alpine:3.21.3

RUN mkdir /gosimplebalancer
RUN touch /gosimplebalancer/config.yaml
COPY --from=build /goapp ./gosimplebalancer

ENTRYPOINT ["/gosimplebalancer/goapp"]

