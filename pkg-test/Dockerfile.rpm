FROM centos:6

ARG package
COPY $package .

RUN rpm -i $package
