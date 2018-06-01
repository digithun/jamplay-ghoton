
FROM golang:1.10.2-stretch as builder
# Copy peer dependency

COPY ./ /go/src/jamplay-ghoton
WORKDIR /go/src/jamplay-ghoton
RUN apt-get update
RUN apt-get install xz-utils
RUN mkdir -p /tmp
RUN make build

FROM acoshift/go-scratch
COPY --from=builder /go/src/jamplay-ghoton/jamplay-ghoton /
COPY ./asset /asset

ENTRYPOINT ["/jamplay-ghoton"]