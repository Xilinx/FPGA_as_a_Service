# Copyright 2018-2022, Xilinx, Inc.
# Copyright 2023, Advanced Micro Device, Inc.
#Author: Brian Xu(brianx@xilinx.com)
#For technical support, please contact k8s_dev@amd.com
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
FROM ubuntu:20.04
RUN mkdir -p /opt/xilinx/k8s-device-plugin/
COPY version_num /opt/xilinx/k8s-device-plugin/
COPY bin/k8s-device-plugin /usr/local/bin/

ENTRYPOINT ["k8s-device-plugin"]
