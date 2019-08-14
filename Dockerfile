FROM golang:1.12 AS gobuild
WORKDIR /go
ADD ./* ./
RUN GOPROXY="https://gocenter.io" go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o rbacbot main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /
COPY --from=gobuild /go/rbacbot /usr/bin/rbacbot
ENTRYPOINT rbacbot