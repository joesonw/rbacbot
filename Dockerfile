FROM golang:1.12 AS gobuild
WORKDIR /go/github.com/joesonw/rbacbot/
ADD . /go/github.com/joesonw/rbacbot/
RUN GOPROXY=https://gocenter.io go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build github.com/joesonw/rbacbot

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /
COPY --from=gobuild /go/github.com/joesonw/rbacbot /bin/rbacbot
ENTRYPOINT /bin/rbacbot