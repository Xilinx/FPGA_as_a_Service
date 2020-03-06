# k8s-fpga-device-plugin
## About
The Xilinx FPGA device plugin for Kubernetes is a Daemonset deployed on the kubernetes(a.k.a k8s) cluster which allows you to:

* Discover the FPGAs inserted in each node of the cluster and expose info of the FPGAs such as quantities, DSA(shell) type and timestamp, etc
* Run FPGA accessible containers in the k8s cluster

More info about k8s device plugin, please refer to https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/device-plugins/

## Prerequisites
* All FPGAs have the DSA(shell) flashed already.
* XRT(version is no older than 2018.3) installed on all worker nodes where there are FPGA(s) inserted
* Container runtime in k8s is docker
* k8s version >= 1.12 (all tests have been running with version 1.12. Old version may or may not work)

## Quick start
Assume there is a running k8s cluster already.

This part shows examples how the FPGA device plugin is deployed and how user APPs are deployed accessing the FPGA(s)

All cmds mentioned in this part run on the master node of k8s cluster. The output of cmds may differ depending on the
yaml file being used.

### Enable Xilinx FPGA support in k8s
#### Deploy FPGA device plugin as daemonset
```
$kubectl create -f fpga-device-plugin.yml
```
#### Check status of daemonset
```
$kubectl get pod -n kube-system

...snippet...

fpga-device-plugin-daemonset-cgq9d   1/1     Running   0          15d
fpga-device-plugin-daemonset-fq689   1/1     Running   0          15d
fpga-device-plugin-daemonset-hmnjr   1/1     Running   0          15d
fpga-device-plugin-daemonset-mkghl   1/1     Running   0          15d

...snippet...
```

Please note, the daemonset will be running on each node of the cluster whether or not there are FPGAs in the node.
If there are FPGAs in the node, logs of the daemonset running there will show something like,

```
$kubectl logs fpga-device-plugin-daemonset-fq689 -n kube-system

time="2019-04-25T18:22:55Z" level=info msg="Starting FS watcher."
time="2019-04-25T18:22:55Z" level=info msg="Starting OS watcher."
time="2019-04-25T18:22:55Z" level=info msg="Starting to serve on /var/lib/kubelet/device-plugins/xilinx_u200_xdma_201820_1-1535712995-fpga.sock"
2019/04/25 18:22:55 grpc: Server.Serve failed to create ServerTransport:  connection error: desc = "transport: write unix /var/lib/kubelet/device-plugins/xilinx_u200_xdma_201820_1-1535712995-fpga.sock->@: write: broken pipe"
time="2019-04-25T18:22:55Z" level=info msg="Registered device plugin with Kubelet xilinx.com/fpga-xilinx_u200_xdma_201820_1-1535712995"
time="2019-04-25T18:22:55Z" level=info msg="Sending 1 device(s) [&Device{ID:1,Health:Healthy,}] to kubelet"
time="2019-04-25T18:32:06Z" level=info msg="Receiving request 1"
time="2019-05-09T18:36:41Z" level=info msg="Receiving request 1"
```
#### Check nodes status and the FPGA resource status on the node

List the nodes in the cluster
```
$kubectl get node

NAME          STATUS   ROLES    AGE   VERSION
fpga-1525-0   Ready    <none>   15d   v1.15.0-alpha.1.109+888d26d1191880-dirty
fpga-u200-0   Ready    <none>   15d   v1.15.0-alpha.1.109+888d26d1191880-dirty
fpga-u200-1   Ready    <none>   15d   v1.15.0-alpha.1.109+888d26d1191880-dirty
fpga-u200-2   Ready    <none>   15d   v1.15.0-alpha.1.109+888d26d1191880-dirty
fpga-u200-3   Ready    <none>   15d   v1.15.0-alpha.1.109+888d26d1191880-dirty
k8smaster     Ready    master   15d   v1.15.0-alpha.1.109+888d26d1191880-dirty
test1         Ready    <none>   15d   v1.15.0-alpha.1.109+888d26d1191880-dirty
test5         Ready    <none>   15d   v1.15.0-alpha.1.109+888d26d1191880-dirty
```

Check FPGA resource in the worker node
```
$kubectl describe node fpga-1525-0

...snippet...

Capacity:
 cpu:                                                    4
 ephemeral-storage:                                      102685624Ki
 hugepages-1Gi:                                          0
 hugepages-2Mi:                                          0
 memory:                                                 16425412Ki
 pods:                                                   110
 xilinx.com/fpga-xilinx_vcu1525_dynamic_5_1-1521279439:  1
Allocatable:
 cpu:                                                    4
 ephemeral-storage:                                      94635070922
 hugepages-1Gi:                                          0
 hugepages-2Mi:                                          0
 memory:                                                 16323012Ki
 pods:                                                   110
 xilinx.com/fpga-xilinx_vcu1525_dynamic_5_1-1521279439:  1

...snippet...

```

### Run jobs accessing FPGA
The Xilinx FPGA resources all have name with following format

xilinx.com/fpga-shell-timestamp

eg. xilinx.com/fpga-xilinx_u200_xdma_201820_1-1535712995

Here, xilinx_u200_xdma_201820_1 is the shell(DSA) version on the FPGA board, and
1535712995 is the timestamp when the shell was built.
```
$date -d @1535712995
Fri Aug 31 03:56:35 PDT 2018
```

The exact name of the FPGA resource on each node can be extracted from the output of
```
$kubectl describe node <node_name>
```

#### Deploy user pod

Here is an example of the yaml file which defines the pod to be deployed.
In the yaml file, the docker image, which has been uploaded to a docker registry, should be specified.
What should be specified as well is, the type and number of FPGA resource being used by the pod.
```
$cat mypod.yaml

apiVersion: v1
kind: Pod
metadata:
  name: my-pod
spec:
  containers:
  - name: my-pod
    image: xilinxatg/fpga-verify:latest
  resources:
    limits:
      xilinx.com/fpga-xilinx_u200_xdma_201820_1-1535712995: 1
  command: ["/bin/sh"]
  args: ["-c", "while true; do echo hello; sleep 5;done;"] 
```

Deploy the pod now

```
$kubectl create -f mypod.yaml
```
#### Check status of the deployed pod
```
$kubectl get pod

...snippet...

my-pod                           1/1     Running   0          7s

...snippet...

```
```
$kubectl describe pod my-pod

...snippet...

Limits:
      xilinx.com/fpga-xilinx_u200_xdma_201820_1-1535712995: 1
    Requests:
      xilinx.com/fpga-xilinx_u200_xdma_201820_1-1535712995: 1

...snippet...

```
#### Run hello world in the pod
```
$kubectl exec -it my-pod /bin/bash
my-pod>source /opt/xilxinx/xrt/setup.sh
my-pod>xbutil scan
my-pod>cd /tmp/alveo-u200/xilinx_u200_xdma_201830_1/test/
my-pod>./verify.exe ./verify.xclbin
```
In this test case, the container image (xilinxatg/fgpa-verify:latest) has been pushed to docker hub. It can be publicly accessed

The image contains verify.xclbin for many types of FPGA, please select the type matching the FPGA resource the pod requests. 

## Known issues
* When there are multiple types of FPGA on one node, the device plugin registers resource for each
  specific typei, but the k8s device plugin framework has issue handling this case. 
  Issue report filed tracking this. https://github.com/kubernetes/kubernetes/issues/70350

## Contact
Brian Xu(brianx@xilinx.com)
