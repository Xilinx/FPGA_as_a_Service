# k8s-fpga-device-plugin
## About
The Xilinx FPGA device plugin for Kubernetes is a [Daemonset](https://kubernetes.io/docs/concepts/workloads/controllers/daemonset/) deployed on the Kubernetes(k8s) cluster which allows you to:

* Discover the FPGAs inserted in each node of the cluster and expose information about FPGA such as number of FPGA, Shell (Target Platform) type and etc.
* Run FPGA accessible containers in the k8s cluster


If you already have an docker and kubernetes environment, you can follow the [Quick Start](https://docs.xilinx.com/r/en-US/Xilinx_Kubernetes_Device_Plugin/Installing-K8s-Device-Plugin-on-Kubernetes) to test k8s-fpga-device-plugin on your own cluster.
You can also check the [Full Tutorial](https://docs.xilinx.com/r/en-US/Xilinx_Kubernetes_Device_Plugin/Build-Kubernetes-Cluster) if you need to build docker, kuberetes cluster environment and test k8s-device-plugin from the beginning.


For detailed information about k8s-device-plugin, Docker and Kubernetes, you can renferece following links:


|Detailed Info               | Description           |
|---------------|-----------------|
| [Kubernetes device plugin](https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/device-plugins/) | Kubernetes community documentation about Kubernetes plugin |
| [Quick Start](https://docs.xilinx.com/r/en-US/Xilinx_Kubernetes_Device_Plugin/Installing-K8s-Device-Plugin-on-Kubernetes) | Quick start on how to install and test k8s-device-plugin |
| [K8s Cluster Full tutorial](https://docs.xilinx.com/r/en-US/Xilinx_Kubernetes_Device_Plugin/Build-Kubernetes-Cluster) | Step by step tutorial starts from install container runtime and Kubernetes cluster |
| [Virtual Device mode](https://docs.xilinx.com/r/en-US/Xilinx_Kubernetes_Device_Plugin/Running-Device-Plugin-in-Virtual-Device-Mode-optional) | Deploy the deivce plugin in Virtual Device mode to allow nyltiple pods to share one single device |
| [Device Name Customization](https://docs.xilinx.com/r/en-US/Xilinx_Kubernetes_Device_Plugin/Device-Name-Customization-for-Kubernetes-Device-plugin-optional) | Customize the device registered name in K8s cluster |
| [FAQ](https://docs.xilinx.com/r/en-US/Xilinx_Kubernetes_Device_Plugin/Support) | Frequently asked questions |

## Prerequisites
* All FPGAs have the Shell(Target Platform) flashed already
* XRT(version is no older than 2018.3) installed on all worker nodes where there are FPGA(s) inserted
* Container runtime in k8s is docker or containerD
* k8s version >= 1.17 (all tests have been running with version 1.17. Old version may or may not work)
* Go 1.18.3 is required if you want to build the device plugin source code

## Contact
Email: k8s_dev@xilinx.com

