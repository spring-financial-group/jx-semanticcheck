FROM alpine:3.14
RUN apk upgrade --no-cache

COPY ./build/ /

ENTRYPOINT ["./jx-semanticcheck", "version"]