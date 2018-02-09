FROM node:8.9.4

ARG STATSD_RELEASE=0.8.0
ARG STATSD_RELEASE_CHECKSUM=ba2c8dcd4f7d863beb26405dd946ac6370a9b012a72c5e575c65c064fa4b4077

ARG STATSD_FILENAME=v${STATSD_RELEASE}.tar.gz
ARG STATSD_RELEASE_URL=https://github.com/etsy/statsd/archive/${STATSD_FILENAME}

# create statsd user
# download statsd release to /usr/local/statsd
# fix bug with recent versions of node
# create default config
RUN set -ex; \
    groupadd -r statsd; useradd -r -g statsd statsd; \
    cd /usr/local; \
    wget $STATSD_RELEASE_URL; \
    echo "$STATSD_RELEASE_CHECKSUM ${STATSD_FILENAME}" | sha256sum --check -; \
    tar xzf ${STATSD_FILENAME}; \
    mv statsd-${STATSD_RELEASE} statsd; \
    sed "s/process.EventEmitter/require('events').EventEmitter/" statsd/lib/config.js -i; \
    rm -f ${STATSD_FILENAME}; \
    mkdir /etc/statsd; \
    cp /usr/local/statsd/exampleConfig.js /etc/statsd/config.js; \
    sed -i 's/graphite.example.com/graphite/' /etc/statsd/config.js; \
    chown -R statsd:statsd /etc/statsd

EXPOSE 8125

USER statsd
WORKDIR /usr/local/statsd
CMD node stats.js /etc/statsd/config.js
