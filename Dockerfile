FROM centos:7

COPY ./build/linux /

ENTRYPOINT ["jx-semanticcheck", "version"]