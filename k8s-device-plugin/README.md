# k8s-fpga-device-plugin
## About
The Xilinx FPGA device plugin for Kubernetes is a [Daemonset]([https://kubernetes.io/docs/concepts/workloads/controllers/daemonset/](https://kubernetes.io/docs/concepts/workloads/controllers/daemonset/)) deployed on the Kubernetes(k8s) cluster which allows you to:

* Discover the FPGAs inserted in each node of the cluster and expose information about FPGA such as number of FPGA, Shell (Target Platform) type and etc.
* Run FPGA accessible containers in the k8s cluster


If you already have an docker and kubernetes environment, you can follow the [Quick Start](https://github.com/Xilinx/FPGA_as_a_Service/blob/master/k8s-device-plugin/quickstart.md) to test k8s-fpga-device-plugin on your own cluster. 
You can check the [Full Tutorial](https://github.com/Xilinx/FPGA_as_a_Service/blob/master/k8s-device-plugin/full-tutorial.md) if you want to build docker, kuberetes cluster environment and test k8s-device-plugin from the beginning.


For detailed information about k8s-device-plugin, Docker and Kubernetes, you can renferece following links:


|Detailed Info               | Description           |
|---------------|-----------------|
| [Kubernetes device plugin](https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/device-plugins/) | Kubernetes community documentation about Kubernetes plugin |
| [Quick Start](https://github.com/Xilinx/FPGA_as_a_Service/blob/master/k8s-device-plugin/quickstart.md) | Quick start on how to install and test k8s-device-plugin |
| [K8S FPGA Device Plugin Full tutorial](https://github.com/Xilinx/FPGA_as_a_Service/blob/master/k8s-device-plugin/full-tutorial.md) | Step by step tutorial starts from install docker and Kubernetes cluster |
| [Kubernetes Docker tutorial](https://github.com/Xilinx/FPGA_as_a_Service/tree/master/k8s-device-plugin/docker) | Build docker image  and test with k8s-fpga-device-plugin |
| [AWS F1 Kubernetes FPGA Plugin](https://github.com/Xilinx/FPGA_as_a_Service/tree/master/k8s-device-plugin/aws) | Install and test k8s-fpga-device-plugin on AWS F1 FPGA |
| [FAQ](https://github.com/Xilinx/FPGA_as_a_Service/blob/master/k8s-device-plugin/FAQ.md) | Frequently asked questions |

## Prerequisites
* All FPGAs have the Shell(Target Platform) flashed already.
* XRT(version is no older than 2018.3) installed on all worker nodes where there are FPGA(s) inserted
* Container runtime in k8s is docker or containerD
* k8s version >= 1.17 (all tests have been running with version 1.17. Old version may or may not work)

## Contact
Email: k8s_dev@xilinx.com
