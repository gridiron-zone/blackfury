FROM ubuntu:22.04

RUN apt-get update && \
    apt-get -y upgrade && \
    apt-get -y install curl jq file

VOLUME [ "/merlion" ]
WORKDIR /merlion
EXPOSE 26656 26657
ENTRYPOINT ["/usr/bin/wrapper.sh"]
CMD ["start"]
STOPSIGNAL SIGTERM

COPY wrapper.sh /usr/bin/wrapper.sh
