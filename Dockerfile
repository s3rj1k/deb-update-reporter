FROM debian:stable-slim
WORKDIR /opt
ADD https://github.com/s3rj1k/deb-update-reporter/releases/download/v1.0.0/deb-update-reporter-linux-x86_64.gz /opt
RUN gzip -d /opt/deb-update-reporter-linux-x86_64.gz
RUN mv /opt/deb-update-reporter-linux-x86_64 /opt/deb-update-reporter
COPY config.yaml /opt
CMD ["/opt/deb-update-reporter", "-update-config", "-config-path=/opt/config.yaml"]
