FROM centos:7.8.2003
MAINTAINER Yu

COPY dbrt_exporter /usr/bin/dbrt_exporter

RUN rm -f /etc/localtime \
&& cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
&& echo 'Asia/Shanghai' > /etc/timezone

CMD ["/bin/bash"]