FROM python:3-slim

RUN pip install --no-cache-dir snmpsim==0.4.7  \
    yamllint==1.25.0 \
    && adduser --system snmpsim

EXPOSE 1024/udp

COPY customer.snmpwalk /data/

USER snmpsim

ENTRYPOINT ["snmpsimd.py"]

CMD ["--data-dir=/data", "--agent-udpv4-endpoint=0.0.0.0:1024"]
