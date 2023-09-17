FROM golang:1.20

WORKDIR /highload-arch

COPY ./ /highload-arch
RUN cd /highload-arch && make build-proj