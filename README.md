# FPGA_as_a_Service
This repository will host FPGA_as_a_Service related projects.

## Contents
| Name                   |  Description |
|------------------------|-------------------|
| [k8s-device-plugin](k8s-device-plugin) | Daemonset deployed on the kubernetes to discover FPGAs inserted in each node and run FPGA accessible containers in the k8s cluster |
| [Xilinx Base Runtime](https://github.com/Xilinx/Xilinx_Base_Runtime) | This project maintains unified Docker images with XRT (Xilinx runtime) preinstalled and provides scripts to setup and flash the Alveo cards. |
| [Containerization](https://github.com/Xilinx/Containerization) | This project provides script to build Docker Application (image) for multiple cloud vendor: Nimbix, AWS and Azure. |
| [Xilinx Container Runtime](https://gitenterprise.xilinx.com/FaaSApps/Xilinx_Container_Runtime) | Xilinx container runtime is an extension of runC, with modification to add xilinx devices before running containers.|
| [XRM](https://github.com/Xilinx/XRM) | XRM - Xilinx FPGA Resource manager is the software to manage all the FPGA hardware on the system. |
