FROM scratch
LABEL maintainer="BigG01 <olig@containerize.ch>"
WORKDIR /bin
COPY bin/main-static /go/bin/main-static
EXPOSE 8080
ENTRYPOINT ["/go/bin/main-static"]