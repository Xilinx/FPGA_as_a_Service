ARG UBUNTU_VERSION

FROM ubuntu:${UBUNTU_VERSION}

RUN apt-get update && apt-get install -y --no-install-recommends \
        wget \
        ca-certificates \
        git \
        build-essential \
        dh-make \
        fakeroot \
        devscripts \
	curl \
        lsb-release && \
    rm -rf /var/lib/apt/lists/*
ARG GOLANG_VERSION
RUN wget -nv -O - "https://storage.googleapis.com/golang/go${GOLANG_VERSION}.linux-$(dpkg --print-architecture).tar.gz" \
    | tar -C /usr/local -xz

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

RUN go get -u golang.org/x/lint/golint

WORKDIR $GOPATH/src/github.com/xilinx/k8s-fpga-device-plugin/
RUN mkdir -p ${GOPATH}/bin

RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

COPY ./src .

RUN ./build

# Second stage
ARG VERSION

FROM ubuntu:16.04

COPY --from=0 /go/src/github.com/xilinx/k8s-fpga-device-plugin/bin/k8s-fpga-device-plugin /usr/local/bin/

ENTRYPOINT ["k8s-fpga-device-plugin"]

