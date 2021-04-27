FROM golang:1.16 as builder
ARG CGO_ENABLED=0
WORKDIR /go/src/github.com/newrelic/nri-snmp
COPY . .
RUN make clean compile

FROM alpine:latest
COPY --from=builder /go/src/github.com/newrelic/nri-snmp/bin /
ADD tests/integration/snmptd/snmp-metrics.yml /snmp-metrics.yml
CMD ["sleep", "1h"]
