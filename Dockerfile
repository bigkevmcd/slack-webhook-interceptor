FROM golang:latest AS build
WORKDIR /go/src
COPY . /go/src
RUN go build -o interceptor ./cmd/interceptor

FROM registry.access.redhat.com/ubi8/ubi-minimal
WORKDIR /root/
COPY --from=build /go/src/interceptor .
EXPOSE 8080
CMD ["./interceptor"]
