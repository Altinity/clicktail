# This builds the binary inside an Alpine Linux container, which is small
FROM alpine:3.3
MAINTAINER Andrey Monakhov <andrey@altinity.com>

# Set us up so we can build the binary
ENV GOROOT /usr/lib/go
ENV GOPATH /gopath
ENV GOBIN /gopath/bin
ENV PATH $PATH:$GOROOT/bin:$GOPATH/bin

WORKDIR /gopath/src/github.com/Altinity/clicktail
COPY . /gopath/src/github.com/Altinity/clicktail/

# Does the package, build, and cleanup as one step to keep size small
RUN apk add --update \
        coreutils \
        go \
        git \
        openssl \
        ca-certificates \
    && ver=$(git rev-parse --short HEAD) \
    && git clean -f \
    && rm -rf .git \
    && go get -ldflags="-X main.BuildID=${ver}" github.com/Altinity/clicktail \
    && apk del git go \
    && rm -rf /var/cache/apk/*

ENV HONEYCOMB_WRITE_KEY NULL
ENV NGINX_LOG_FORMAT_NAME combined
ENV NGINX_CONF /etc/nginx.conf
ENV HONEYCOMB_SAMPLE_RATE 1
ENV NGINX_ACCESS_LOG_FILENAME access.log

CMD [ "/bin/sh", "-c", "clicktail \
            --parser nginx \
            --file /var/log/nginx/$NGINX_ACCESS_LOG_FILENAME \
            --dataset clicktail.nginx_logs \
            --samplerate $HONEYCOMB_SAMPLE_RATE \
            --nginx.conf $NGINX_CONF \
            --nginx.format $NGINX_LOG_FORMAT_NAME \
            --tail.read_from end" ]
