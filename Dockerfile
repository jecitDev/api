FROM golang:1.16.3 AS builder
ADD . /go/src/api
WORKDIR /go/src/api
RUN go get . && \
    go install && \
    go build -o engine main.go
EXPOSE 443

FROM alpine:edge AS runtime
LABEL app="api" version="1.13" maintainer="Zico Alamsyah"
ENV serverport="443" dbapi="80.167" dbapp="MSSQL" 
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2
WORKDIR /go/src/api
EXPOSE 443
COPY --from=builder /go/src/api/engine /go/src/api
CMD ./engine 