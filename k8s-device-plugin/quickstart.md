## Quick start
Assume there is a running k8s cluster already.

This part shows examples how the FPGA device plugin is deployed and how user APPs are deployed accessing the FPGA(s)

All cmds mentioned in this part run on the master node of k8s cluster. The output of cmds may differ depending on the yaml file being used.

### Enable Xilinx FPGA support in k8s

#### Remove existing device plugin daemonset
 If you already deployed the device plugin on your cluster, before migrating to version 1.1.0+ or updating the env config value in yaml file, you need to remove the device plugin daemonset and all pods you created with FPGA resources allocated.
```
Check existing device plugin daemonset
#kubectl get daemonset -n kube-system

Remove existing device plugin daemonset
- device plugin version 1.1.0 and previous :
#kubectl delete daemonset fpga-device-plugin-daemonset -n kube-system

- device plugin version 1.1.0+ :
#kubectl delete daemonset device-plugin-daemonset -n kube-system

Check all created user pods :
#kubectl get pod

Check user pod allocated resrouces :
#kubectl describe pod <pod-name>

Delete Pod with FPGA(Alveo) devices allocated :
#kubectl delete pod <pod-name>
```

####  Config yaml file (Only required for Alveo U30 device and AWS VT1 node)

In yaml file `fpga-device-plugin.yml`, the following 2 environmental variables define the U30 naming convention and access granularity:
```
...
      containers:
      - image: public.ecr.aws/xilinx_dcg/k8s-device-plugin:1.1.0
        name: device-plugin
        env:
        - name: U30NameConvention
          value: "{CommonName|ExactName}"
        - name: U30AllocUnit
          value: "{Card|Device}"
...
```
U30NameConvention sets the pod resource limit field to be one of the following:

ExactName: The U30 FPGA devices will be registered in format `amd.com/xilinx_u30_gen3x4_base_GA_VERSION-timestamp`, where GA_VERSION refers to General Release (GA) 1,2 or 3.

CommonName: The U30 FPGA devices as `amd.com/ama_u30`, where both forward and backward compatibility.

U30AllocUnit in pod resource limit numeration will have units of either Card or Device.

If any of the input values for U30NameConvention/U30AllocUnit is empty or invalid, device plugin will set the invalid input as the default value CommonName/Card.

**Attention:** If you plan to use Device as U30AllocUnit under on prem(bare-meatl) k8s cluster, the 2 same board U30 devices will be able to be assigned into separate pods, running `xbutil reset` on one of the U30 device will leading the other same board U30 device be reset. Please be careful while running 'xbutil reset' command.

#### Deploy FPGA device plugin as daemonset
```
$kubectl create -f fpga-device-plugin.yml
```
#### Check status of daemonset pod
```
$kubectl get pod -n kube-system

...snippet...

device-plugin-daemonset-cgq9d   1/1     Running   0          15d
device-plugin-daemonset-fq689   1/1     Running   0          15d
device-plugin-daemonset-hmnjr   1/1     Running   0          15d
device-plugin-daemonset-mkghl   1/1     Running   0          15d

...snippet...
```

Please note, the daemonset will be running on each node of the cluster whether or not there are FPGAs in the node.
If there are FPGAs in the node, logs of the daemonset running there will show as follow,

```
$kubectl logs fpga-device-plugin-daemonset-fq689 -n kube-system

time="2022-08-12T16:44:25Z" level=info msg="Plugin Version: 1.1.0"
time="2022-08-12T16:44:25Z" level=info msg="Set U30NameConvention: ExactName"
time="2022-08-12T16:44:25Z" level=info msg="Set U30 AllocUnit: Device"
time="2022-08-12T16:44:25Z" level=info msg="Starting FS watcher."
time="2022-08-12T16:44:25Z" level=info msg="Starting OS watcher."
time="2022-08-12T16:44:25Z" level=info msg="Starting to serve on /var/lib/kubelet/device-plugins/xilinx_u50_gen3x16_xdma_201920_3-0-fpga.sock"
2022/08/12 16:44:25 transport: http2Server.HandleStreams failed to read frame: read unix /var/lib/kubelet/device-plugins/xilinx_u50_gen3x16_xdma_201920_3-0-fpga.sock->@: read: connection reset by peer
time="2022-08-12T16:44:25Z" level=info msg="Registered device plugin with Kubelet amd.com/xilinx_u50_gen3x16_xdma_201920_3-0"
time="2022-08-12T16:44:25Z" level=info msg="Check SeialNums arry: [5001A698P02E]"
time="2022-08-12T16:44:25Z" level=info msg="Sending 1 device(s) [&Device{ID:0000:82:00.1,Health:Healthy,}] to kubelet"
```
#### Check nodes status and the FPGA resource status on the node

List the nodes in the cluster
```
$kubectl get node

NAME          STATUS   ROLES    AGE   VERSION
fpga-u50-0    Ready    <none>   15d   v1.22.2
fpga-u200-0   Ready    <none>   15d   v1.22.2
fpga-u200-1   Ready    <none>   15d   v1.22.2
fpga-u200-2   Ready    <none>   15d   v1.22.2
fpga-u200-3   Ready    <none>   15d   v1.22.2
k8smaster     Ready    master   15d   v1.22.2
```

Check FPGA resource in the worker node
```
$kubectl describe node fpga-u200-0

...snippet...

Capacity:
  amd.com/xilinx_u200_xdma_201830_1-1542252769:                       1
  cpu:                                                                12
  ephemeral-storage:                                                  362372628Ki
  hugepages-1Gi:                                                      0
  hugepages-2Mi:                                                      0
  memory:                                                             98904344Ki
  pods:                                                               110
Allocatable:
  amd.com/xilinx_u200_xdma_201830_1-1542252769:                       1
  cpu:                                                                12
  ephemeral-storage:                                                  333962613412
  hugepages-1Gi:                                                      0
  hugepages-2Mi:                                                      0
  memory:                                                             98801944Ki
  pods:                                                               110

...snippet...

```

### Run jobs accessing FPGA

The exact name of the FPGA resource on each node can be extracted from the output of
```
$kubectl describe node <node_name>
```

#### Deploy user pod

Here is an example of the yaml file which defines the pod to be deployed.
In the yaml file, the docker image, which has been uploaded to a docker registry, should be specified.
What should be specified as well is, the type and number of FPGA resource being used by the pod.
```
$cat dp-pod.yaml

apiVersion: v1
kind: Pod
metadata:
  name: mypod
spec:
  containers:
  - name: mypod
    image: xilinxatg/fpga-verify:latest
    resources:
      limits:
        amd.com/xilinx_u200_xdma_201830_1-1542252769: 1
    command: ["/bin/sh"]
    args: ["-c", "while true; do echo hello; sleep 10;done"]
```

Deploy the pod

```
$kubectl create -f dp-pod.yaml
```
#### Check status of the deployed pod
```
$kubectl get pod

...snippet...

mypod                           1/1     Running   0          7s

...snippet...

```
```
$kubectl describe pod mypod

...snippet...

Limits:
      xilinx.com/fpga-xilinx_u200_xdma_201820_1-1535712995: 1
    Requests:
      xilinx.com/fpga-xilinx_u200_xdma_201820_1-1535712995: 1

...snippet...

```
#### Run hello world in the pod
```
$kubectl exec -it mypod /bin/bash
my-pod>source /opt/xilinx/xrt/setup.sh
my-pod>xbutil scan
my-pod>cd /tmp/alveo-u200/xilinx_u200_xdma_201830_1/test/
my-pod>./validate.exe ./verify.xclbin
```
In this test case, the container image (xilinxatg/fgpa-verify:latest) has been pushed to docker hub. It can be publicly accessed

The image contains verify.xclbin for many types of FPGA, please select the type matching the FPGA resource the pod requests.






