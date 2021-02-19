FROM python:3-slim AS base

RUN pip install --no-cache-dir snmpsim==0.4.7  \
    yamllint==1.25.0 \
    && adduser --system snmpsim

EXPOSE 1024/udp
USER snmpsim

ENTRYPOINT ["snmpsimd.py"]
CMD ["--agent-udpv4-endpoint=0.0.0.0:1024"]

FROM base AS toubleshooting

ARG SNMPWAL_FILE=toubleshooting.snmpwalk
COPY ${SNMPWAL_FILE} /data/

CMD ["--data-dir=/data", "--agent-udpv4-endpoint=0.0.0.0:1024"]
