#Copyright 2018-2021 Xilinx Corporation. All Rights Reserved.
##FPGA as a Service (k8s_dev@xilinx.com)
#
#Licensed under the Apache License, Version 2.0 (the "License");
#you may not use this file except in compliance with the License.
#You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
#Unless required by applicable law or agreed to in writing, software
#distributed under the License is distributed on an "AS IS" BASIS,
#WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#See the License for the specific language governing permissions and
#limitations under the License.
FROM golang:latest as builder

# Define working directory.
WORKDIR /opt/xilinx-device-plugin

COPY . .
RUN	go env -w GO111MODULE=auto
RUN ./build
    
FROM debian:bullseye-slim

LABEL maintainer="k8s_dev@xilinx.com <k8s_dev@xilinx.com>"


COPY --from=builder /opt/xilinx-device-plugin/bin /usr/local/bin
ENTRYPOINT ["/usr/local/bin/xilinx-device-plugin"]