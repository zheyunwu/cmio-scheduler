FROM debian:stretch-slim

WORKDIR /

COPY cmio-scheduler /usr/local/bin

CMD ["cmio-scheduler"]
