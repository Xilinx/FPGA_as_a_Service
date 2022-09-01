# AWS Kubernetes FPGA Plugin

## Note

The FPGA plugin in this folder is only for AWS running XRT prior 2019.2. In XRT 2019.2,
the normal plugin in the parent folder works also for AWS, and there is no limitation
mentioned below for this plugin.

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

Once the kubernetes cluster and XRT is being setup on aws, you can follow following step to install and test the kubernetes plugin.

### Step 1: Deploy the plugin as daemonset
Download plugin source:
```
#git clone  https://github.com/Xilinx/FPGA_as_a_Service.git
```
Deploy FPGA device plugin as daemonset:  
```
#kubectl create -f ./FPGA_as_a_Service/k8s-fpga-device-plugin/trunk/fpga-device-plugin.yml 
``` 
To check the status of daemonset:  
```
#kubectl get pod -n kube-system  
```
Get node name:  
```
#kubectl get node  
```
Check FPGA resource in the worker node:  
```
#kubectl describe node nodename  
```
You should get the FPGA resources name under the pods information.

### Step 2: deploy a user pod
```
#kubectl create -f mypod.yaml
```

For more details, please refer to the README file of the Xilinx FPGA plugin
### Step 3: Run the test in pod
After user pod status turns to Running, run hello world in the pod:  
```
#kubectl exec -it my-pod /bin/bash  
#my-pod>source /opt/xilinx/xrt/setup.sh  
#my-pod>export INTERNAL_BUILD=1  
#my-pod>xbutil scan  
#my-pod>cd /opt/test/  
#my-pod>./helloworld vector_addition_hw.awsxclbin
```
**Note:**  Need to set the INTERNAL_BUILD=1 if xbutil complain the version not match:  

## Build plugin binary

```
#./build
```

The output is the binary 'aws-fpga-device-plugin' in the current folder

