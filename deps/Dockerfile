FROM scratch

ENV GIT_HASH=<GIT_HASH>

ADD rootfs.tar.gz /
COPY cryodns /
COPY config.gcfg /
COPY run.sh /

CMD ["/run.sh"]
