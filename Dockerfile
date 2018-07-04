FROM alpine:3.7
ARG GOOS
ARG GOARCH
ARG BUILD
ARG TAG
RUN apk add -U --no-cache ca-certificates
COPY cmd/stellar/stellar /bin/stellar
COPY cmd/sctl/sctl /bin/sctl
ENTRYPOINT ["/bin/stellar"]
EXPOSE 7946 9000
CMD ["-h"]
