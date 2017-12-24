FROM armv7/armhf-ubuntu:14.04
MAINTAINER Nani Shin <nani.shin@gmail.com>

ENV QEMU_EXECVE 1

ADD local_files/docker_init_files.tar.gz /

RUN [ "cross-build-start" ]
RUN if [ -x /debootstrap/debootstrap ] ; then /debootstrap/debootstrap --second-stage ; fi \
 && dpkg-divert --local --rename --add /sbin/initctl

RUN apt-get update \
 && apt-get install -y supervisor \
 && apt-get install -y ntpdate \
 && apt-get install -y alsa-utils libao4 libao-common libid3tag0 \
 && apt-get install -y libmad0 \
 && apt-get install -y mpg321 \
 && apt-get install -y tzdata \
 && apt-get -f install \
 && apt-get clean -y && apt-get autoclean -y && apt-get autoremove -y

RUN mkdir -p /var/lib/resolvconf /var/log/supervisor \
 && touch /var/lib/resolvconf/linkified \
 && echo 'resolvconf resolvconf/linkify-resolvconf boolean false' | debconf-set-selections
RUN [ "cross-build-end" ]

EXPOSE 7890 8080

CMD ["/usr/bin/start-upcycling-xmas-tree"]
