FROM phusion/baseimage:0.9.22
MAINTAINER Nicholas Evans <nevans@410labs.com>

# choose a timezone at build-time
# use `--build-arg CONTAINER_TIMEZONE=Europe/Brussels` in `docker build`
ARG CONTAINER_TIMEZONE="America/New_York"
ENV DEBIAN_FRONTEND noninteractive
ENV PATH /usr/local/bin:$PATH

RUN set -ex; \
  apt-get -y update; \
  apt-get -y upgrade; \
  apt-get -y install \
    build-essential \
    expect \
    libcairo2 \
    libcairo2-dev \
    libffi-dev \
    pkg-config \
    python-cairo \
    python-dev \
    python-flup \
    python-ldap \
    python-pip \
    python-rrdtool \
    sqlite3 \
  ; \
  rm -rf /var/lib/apt/lists/*; \
  if [ ! -z "${CONTAINER_TIMEZONE}" ]; then \
    ln -sf /usr/share/zoneinfo/$CONTAINER_TIMEZONE /etc/localtime && \
    dpkg-reconfigure -f noninteractive tzdata; \
  fi

ENV PYTHONPATH="/opt/graphite/lib/:/opt/graphite/webapp/"
ARG django_version=1.8.18
ARG gunicorn_version=19.7.1
ARG version=1.1.1
ARG whisper_version=${version}
ARG carbon_version=${version}
ARG graphite_version=${version}
ARG graphite_github=https://github.com/graphite-project

RUN set -ex; \
  pip install --upgrade pip; \
  pip install django==$django_version; \
  pip install gunicorn==$gunicorn_version; \
  mkdir -p /opt/graphite; \
  pip install $graphite_github/whisper/tarball/$whisper_version; \
  pip install $graphite_github/carbon/tarball/$carbon_version; \
  pip install $graphite_github/graphite-web/tarball/$graphite_version; \
  cd /opt/graphite/conf; \
  rename 's/\.conf\.example$/\.conf/' *.conf.example

## config graphite
ADD conf/opt/graphite/conf/*.conf /opt/graphite/conf/
ADD conf/opt/graphite/webapp/graphite/local_settings.py /opt/graphite/webapp/graphite/local_settings.py
ADD conf/usr/local/bin/django_admin_init.exp /usr/local/bin/django_admin_init.exp
ADD conf/usr/local/bin/manage.sh /usr/local/bin/manage.sh
WORKDIR /opt/graphite/webapp

RUN set -ex; \
  mkdir -p /var/log/graphite/; \
  django-admin.py collectstatic --noinput --settings=graphite.settings; \
  django_admin_init.exp

EXPOSE 8080
VOLUME ["/opt/graphite/storage", "/var/log"]
WORKDIR /opt/graphite
CMD gunicorn wsgi --workers=4 --bind=0.0.0.0:8080 --preload \
  --log-file=- --access-logfile=- \
  --pythonpath=/opt/graphite/webapp/graphite
