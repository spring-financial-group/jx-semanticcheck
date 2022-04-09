FROM alpine:3.14
RUN apk upgrade --no-cache
RUN apk add git

COPY ./build/ /

ENTRYPOINT ["./jx-semanticcheck", "version"]