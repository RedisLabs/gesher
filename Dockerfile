FROM  golang:1.14-alpine as stage1
ENV GOPATH=/go
WORKDIR /go/src/github.com/RedisLabs/gesher
RUN mkdir -p /go/src/github.com/RedisLabs/gesher
RUN apk add --update \
        bash \
        zip \
        ca-certificates \
        make \
        build-base \
        curl \
        git
RUN curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | bash -s -- -b $GOPATH/bin v1.18.0
COPY / ./
RUN golangci-lint run -E gofmt --deadline 10m
RUN go build ./cmd/manager

FROM scratch
COPY --from=stage1 /go/src/github.com/RedisLabs/gesher/manager /
