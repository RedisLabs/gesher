FROM  golang:1.16-alpine as stage1
ARG version=dev
LABEL name="redislabs/gesher" \
      maintainer="support@redislabs.com" \
      vendor="Redis Labs" \
      version="${version}" \
      release="1"
ENV GOPATH=/go
WORKDIR /go/src/github.com/RedisLabs/gesher
RUN mkdir -p /go/src/github.com/RedisLabs/gesher
COPY / ./
RUN CGO_ENABLED=0 go build -tags netgo -ldflags '-w -extldflags "-static"' ./cmd/manager

FROM scratch
COPY --from=stage1 /go/src/github.com/RedisLabs/gesher/manager /
ENTRYPOINT ["/manager"]
