# aws_k8s_fpga_plugin

## About

This kubernetes FPGA device plugin is created for cluster running on aws F1 nodes.

The FPGA cards in F1 are different to those Xilinx ones. The differences are in both hardware
and software, so the plugin created for Xilinx FPGA & XRT doesn't work for AWS. 

In this dedicated plugin,

* Shell version is hard-coded in XRT for aws, and it is not exported to sysfs. so the
plugin also hard-codes the shell version as "xilinx_aws-vu9p-f1-04261818_dynamic_5_0"
* Shell timestamp is not being used. To keep the same FPGA resource format, the shell
timestamp info is set as "0" 

## Limitations on user pods

The plugin itself only relies on a readonly sysfs, so the same to the plugin for Xilinx
FPGAs, this aws plugin is also deployed as normal container in which the sysfs is mounted
as readonly. While for user pods, in order to run awssak scan/list and/or aws specific 
fpga- commands, root access has to be granted within the container, which means, in addtion
to the install of xrt/aws-xrt and aws fpga tools from https://github.com/aws/aws-fpga.git,
the user pods have to be deployed in 'privileged' mode.

## Quick start

Once the kubernetes cluster is setup,

### Deploy the plugin as daemonset
```
$kubectl create -f aws-fpga-device-plugin.yaml
```
### Example to deploy a user pod
```
$kubectl create -f mypod.yaml
```

More details, please refer to the README file of the Xilinx FPGA plugin

## Build plugin binary

```
#./build
```

The output is the binary 'aws-fpga-device-plugin' in the current folder

## Build plugin docker image

```
#docker build -t name_of_the_image .
```
There is a docker image built already and pushed to dockerhub

```
xilinxatg/aws_k8s_fpga_plugin:06272019
```
