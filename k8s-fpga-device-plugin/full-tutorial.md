# Xilinx FPGA Plugin Deployment Full Tutorial

This documentation describes how to deploy FPGA plugin with Docker and Kubernetes on RedHat, CentOS and Ubuntu.
## 1. Install Docker

### 1.1 Prerequisites

CentOS:

-   A maintained/supported version of CentOS 7+
-   A user account with sudo privileges
-   Terminal access
-   CentOS Extras repository – this is enabled by default, but if yours has been disabled you’ll need to re-enable it
-   Software package installer yum

RedHat:

-   A maintained/supported version of Redhat 7+ (We tested with RHEL 7.8)
-   A user account with sudo privileges
-   Terminal access
-   Software package installer yum

Ubuntu:

-   Ubuntu operating system 16.04+
-   A user account with sudo privileges
-   Command-line/terminal
-   Docker software repositories (optional)
### 1.2 Install Docker on CentOS 7 / RedHat 7.8 With Yum

#### Step 1: Update Docker Package Database

`#sudo yum check-update`

#### Step 2: Install the Dependencies

`#sudo yum install -y yum-utils device-mapper-persistent-data lvm2`

#### Step 3: Add the Docker Repository to CentOS / Redhat

To install the  **edge**  or  **test**  versions of Docker, you need to add the Docker CE stable repository to your system. To do so, run the command:

`#sudo yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo`

A **stable** release is tested more thoroughly and has a slower update cycle. On the other hand, **Edge** release updates are more frequent but aren’t subject to as many stability tests.

**Note:** If you’re only going to use the stable release, don’t enable these extra repositories. The Docker installation process defaults to the latest version of Docker unless you specify otherwise. Leaving the stable repository enabled makes sure that you aren’t accidentally updating from a stable release to an edge release.

#### Step 4: Install Docker On CentOS/Redhat Using Yum

With everything set, you can finally move on to installing Docker on CentOS 7 by running:

`#sudo yum install docker`

The system should begin the installation. Once it finishes, it will notify you the installation is complete and which version of Docker is now running on your system.

Your operating system may ask you to accept the GPG key. This is like a digital fingerprint, so you know whether to trust the installation.

#### Step 5: Manage Docker Service

Although you have installed Docker on CentOS, the service is still not running.

To start the service, enable it to run at startup. Run the following commands in the order listed below.

Start Docker:

`#sudo systemctl start docker`

Enable Docker:

`#sudo systemctl enable docker`

Check the status of the service:

`#sudo docker run hello-world`

### 1.3 Install Docker on Ubuntu With Apt-get

#### Step 1: Update Software Repositories

`#sudo apt-get update`

#### Step 2: Install Docker

`#sudo apt-get install docker.io`

#### Step 3: Manage Docker Service

Although you have installed Docker on Ubuntu, the service is still not running.

To start the service, enable it to run at startup. Run the following commands in the order listed below.

Start Docker:

`#sudo systemctl start docker`

Enable Docker:

`#sudo systemctl enable docker`

Check the status of the service:

`#sudo docker run hello-world`

## 2. Install Kubernetes

You will install these packages on all of your machines:

-   `kubeadm`: the command to bootstrap the cluster.

-   `kubelet`: the component that runs on all of the machines in your cluster and does things like starting pods and containers.

-   `kubectl`: the command line util to talk to your cluster.

Here is the referred document from Kubernetes:

[https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/](https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/)

### 2.1 Install kubeadm, kubelet and kubectl on CentOS / Redhat

#### Step 1: Set Kubernetes repo

`#update-alternatives --set iptables /usr/sbin/iptables-legacy`

`#cat /etc/yum.repos.d/kubernetes.repo`  

```
[kubernetes]  
name=Kubernetes  
baseurl=[https://packages.cloud.google.com/yum/repos/kubernetes-el7-x86_64](https://packages.cloud.google.com/yum/repos/kubernetes-el7-x86_64)  
enabled=1  
gpgcheck=1  
repo_gpgcheck=1  
gpgkey=https://packages.cloud.google.com/yum/doc/yum-key.gpg https://packages.cloud.google.com/yum/doc/rpm-package-key.gpg
```

#### Step 2: Set SELinux in permissive mode (effectively disabling it)

```
#sudo setence 0
#sudo sed -i 's/^SELINUX=enforcing$/SELINUX=permissive/' /etc/selinux/config 
```

**Note**: Setting SELinux in permissive mode by running setenforce 0 and sed ... effectively disables it. This is required to allow containers to access the host filesystem, which is needed by pod networks for example. You have to do this until SELinux support is improved in the kubelet.  

Some users on RHEL/CentOS 7 have reported issues with traffic being routed incorrectly due to iptables being bypassed. You should ensure net.bridge.bridge-nf-call-iptables is set to 1 in your sysctl config, e.g.

`#cat /etc/sysctl.d/k8s.conf  `

net.bridge.bridge-nf-call-iptables = 1  

`#sudo sysctl --system  `

Make sure that the br_netfilter module is loaded before this step. This can be done by running

`#lsmod | grep br_netfilter`

To load it explicitly call

`#sudo modprobe br_netfilter`

#### Step 3: Install Kubernetes
```
#sudo yum install -y kubelet kubeadm kubectl --disableexcludes=kubernetes`
#sudo systemctl enable --now kubelet
```

### 2.2 Install kubeadm, kubelet and kubectl on Ubuntu

#### Step 1: Set kubernetes repo
```
#sudo apt-get update
#sudo apt-get install -y iptables arptables ebtable
```
#### Step 2: Install Kubernetes

```bash
#sudo apt-get update && sudo apt-get install -y apt-transport-https curl
#curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add -
#sudo apt-get update
#sudo apt-get install -y kubelet kubeadm kubectl
#sudo apt-mark hold kubelet kubeadm kubectl
```

**Note**:  
If you want to install specified version of kubelet, kubeadm and kubectl. You can use following command to check and install available versions.

For Redhat: 
```
#sudo yum install -y kubelet-1.18.9 kubeadm-1.18.9 kubectl-1.18.9 --disableexcludes=kubernetes
#sudo systemctl enable --now kubelet
```
For Ubuntu:
```
#sudo apt-cache policy kubeadm
#sudo apt-get install -y kubelet=1.18.9-00 kubeadm=1.18.9-00 kubectl=1.18.9-00
#sudo apt-mark hold kubelet kubeadm kubectl
```
## 3. Configure Cluster


### 3.1 Disable swap (this step need to be done on all your nodes)

`#sudo swapoff -a`

**Note**: If there is no enough space on system, please try disable swap and remove the swap file.

This command only temporary disable swap, run this command each time after reboot the machine.

### 3.2 Build Master Node

#### Step 1: Init master node
```
#sudo kubeadm init --pod-network-cidr=10.244.0.0/16
#sudo mkdir -p $HOME/.kube
#sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
#sudo chown $(id -u):$(id -g) $HOME/.kube/config
```

**Note:** For issues like: "The connection to the server localhost:8080 was refused - did you specify the right host or port?"

Check port status:

`#netstat -nltp | grep apiserver`

Adding environment variable in ~/.bash_profile

`#export KUBECONFIG=/etc/kubernetes/admin.conf`

`#source ~/.bash_profile`

#### Step 2: configure flannel

install flannel (for Kubernetes version 1.7+)  

`#sysctl net.bridge.bridge-nf-call-iptables=1  `

`#kubectl apply -f https://raw.githubusercontent.com/coreos/flannel/master/Documentation/kube-flannel.yml`

For other version, please refer [https://github.com/coreos/flannel](https://github.com/coreos/flannel) to do the configuration.

#### Step 3: check the pod

`#sudo kubectl get pod -n kube-system -o wide`

### 3.3 Adding worker node (slave node)

If there are multiple server machines or AWS instances you want to add into cluster as worker node, you need to follow our perivous step install matched version of kubectl kubeadm and kubelet. 

Login to your master node, use following command to get your token command for joining cluster:

`kubeadm token create --print-join-command`

You will get a output command like :

`kubeadm join 192.168.54.128:6443 --token mg4o13.4ilr1oi605tj850w   --discovery-token-ca-cert-hash sha256:363b5b8525ddb86f4dc157f059e40c864223add26ef53d0cfc9becc3cbae8ad3`

Insert the output command on the worker nodes you want to add into cluster. Then checking adding result on master node with command:

`kubectl get node`

### 3.4 Labeling your node (Optional)
If you want to create a pod that gets scheduled to you chosen node, you need to label the node first and setting nodeSelector in your pod yaml file. 

Choose your nodes, and add a label to it:
`kubectl label nodes <your-node-name> disktype=ssd`

Verify that your chosen node has a `disktype=ssd` label:
`kubectl get nodes --show-labels`

The output is similar to this:
```
NAME      STATUS    ROLES    AGE     VERSION        LABELS
Master    Ready     master   1d      v1.18.9        ...,kubernetes.io/hostname=master
worker0   Ready     <none>   1d      v1.18.9        ...,disktype=ssd,kubernetes.io/hostname=worker0
worker1   Ready     <none>   1d      v1.18.9        ...,kubernetes.io/hostname=worker1
worker2   Ready     <none>   1d      v1.18.9        ...,kubernetes.io/hostname=worker2
```
This example pod configuration yaml file describes a pod that has a node selector, disktype: ssd. This means that the pod will get scheduled on a node that has a disktype=ssd label.
```
apiVersion: v1
kind: Pod
metadata:
  name: nginx
  labels:
    env: test
spec:
  containers:
  - name: nginx
    image: nginx
    imagePullPolicy: IfNotPresent
  nodeSelector:
    disktype: ssd
```
## 4. Install Xilinx Runtime (XRT)

### 4.1 Install with XRT package
For bare-metal machine you can directly install xrt packages with "sudo apt install xrt_version.deb" or "sudo yum install xrt_version.rpm". 
XRT installation tutorial: https://xilinx.github.io/XRT/master/html/install.html

### 4.2 Build XRT from source code on AWS F1 (optional)
Here we will introduce how to build and install XRT on an AWS F1 CentOS server. 
We will download XRT from github, build and install it with following command line.
#### 4.2.1 Setup tool

`#scl enable devtoolset-6 bash`

If scl and devtoolset is not installed, then need to install the listed tools.

#### 4.2.2 Setup AWS FPGA

Here need to download aws FPGA because XRT build will depend on the it.

`#git clone http://github.com/aws/aws-fpga.git`

`#export AWS_FPGA_REPO_DIR="path of aws-fpga"`

#### 4.2.3 Build and Install XRT

**Note:** Based on current test, XRT 2019.2.0.3 works well on AWS F1, the master version has issue on some F1 instance. So here we recommend to use 2019.2.0.3 version.

#### Step 1: Build XRT
```
#git clone -b 2019.2.0.3 https://github.com/Xilinx/XRT.git
#./src/runtime_src/tools/scripts/xrtdeps.sh
#cd build
#./build.sh
#cd Release
#make package
```
**Note**: Need to make sure $AWS_FPGA_REPO_DIR is set to the right directory of aws-fpga before running build.

#### Step 2: Install XRT

`#yum install xrt_201920.2.3.0_7.7.1908-xrt.rpm`
`#yum install xrt_201920.2.3.0_7.7.1908-aws.rpm`

Please refer to the full instruction on how to build and install XRT:

[https://github.com/Xilinx/XRT/blob/master/src/runtime_src/doc/toc/build.rst](https://github.com/Xilinx/XRT/blob/master/src/runtime_src/doc/toc/build.rst)

`#source /opt/xilinx/xrt/setup.sh`

To check the FPGA device on the system:

```
#systemctl start mpd
#systemctl status mpd
#xbutil scan
```

## 5. Install Kubernetes FPGA Plugin

If you only have one node (master) in your cluster or plan to deploy pod on master node, to enable this configuration, we need to configure the control plane node.

### 5.1 Control plane node isolation

By default, your cluster will not schedule Pods on the control-plane node for security reasons. If you want to be able to schedule Pods on the control-plane node, e.g. for a single-machine Kubernetes cluster for development, run:

`#kubectl taint nodes --all node-role.kubernetes.io/master-`

### 5.2 Install Kubernetes FPGA plugin

Following steps need to be done on your master node. After install FPGA plugin you don't need to do any other configuration when adding new nodes into the cluster.

####   Step 1: Download plugin source

`#git clone  https://github.com/Xilinx/FPGA_as_a_Service.git`

Deploy FPGA device plugin as daemonset:  

`#kubectl create -f ./FPGA_as_a_Service/k8s-fpga-device-plugin/fpga-device-plugin.yml `

To check the status of daemonset:  

`#kubectl get daemonset -n kube-system  `

Get node name:  

`#kubectl get node  `

Check FPGA resource in the worker node:  

`#kubectl describe node nodename  `

You should get the FPGA resources name under the pods information.

## 6 Deploy user pod
### 6.1 Edit your pod creating yaml file
Here we use following yaml file as an example, you can found it under ./exiaws/mypod.yaml
```
apiVersion: v1
kind: Pod
metadata:
  name: my-pod
spec:
  containers:
  - name: my-pod
    image: centos:bx #user needs to build use its own docker image
    securityContext:
      privileged: true
    resources:
      limits:
        xilinx.com/fpga-xilinx_aws-vu9p-f1-04261818_dynamic_5_0-0: 1
    command: ["/bin/sh"]
    args: ["-c", "while true; do echo hello; sleep 10;done"]
    volumeMounts:
      - name: sys
        mountPath: /sys
  volumes:
    - name: sys
      hostPath:
        path: /sys
```
You need to do following configuration before  you creating the pod:
1) Modify the image to be "xilinxatg/aws-fpga-verify:20200131", you can change this to your own docker images. We will introduce how to build your own docker image at section 7 of this tutorial.
2) You can use `kubectl describe node [nodename]` on master node to check available resources' number  and names.
3) Modify the resources: set limits same as the that in worker node like "xilinx.com/fpga-xilinx_aws-vu9p-f1_dynamic_5_0-43981: 1". 



### 6.2 Create pod
Create pod from yaml file: `#kubectl create -f mypod.yaml`
To check status of the deployed pod: `#kubectl get pod`
```
NAME     READY   STATUS    RESTARTS   AGE
my-pod   1/1     Running   0          59m
```
If the pod is stuck at container-creating step or being evicted, use `#kubectl describe pod my-pod` to check detailed information about pod creating process.
### 6.3 Validate pod
After the pod status turns to Running, run hello world in the pod:  

`#kubectl exec -it my-pod -- /bin/bash  `

`#my-pod>source /opt/xilinx/xrt/setup.sh  `

**Note:**  Need to set the INTERNAL_BUILD=1 if xbutil complain the version not match inside pod:  
```
#my-pod>export INTERNAL_BUILD=1  
#my-pod>xbutil scan  
#my-pod>cd /opt/test/  
#my-pod>./helloworld vector_addition_hw.awsxclbin
```
## 7. How to build new docker image

We will use an example to explain how a new docker image with desired contents such as your xclbin, your host code etc. can be built. Please note that any accelerator (FPGA) docker image should be derived form the base docker Xilinx image **xilinxatg/aws-fpga-verify:20200131** already hosted at the Docker Hub.

### 7.1 Prerequisites

To host a docker image, you need some sort of service. You can host it locally if you like (please read online docker instructions for that). However, this instruction uses [Docker Hub](https://hub.docker.com/)  as the hosting service.

Go to [https://hub.docker.com/signup](https://hub.docker.com/signup) to create a Docker Hub account (if you do not have one already), and then create a docker repository.

In this document, we are using an example docker account named as **xilinxatg**  and an example repository named as **k8s-plugin-dev**. Please substitute these by your account and repository names respectively. Also, please note that you can set your repository as private if you do not want others to see it.

### 7.2 Prepare docker images

####   Step 1: Login to your Docker Hub account

`#docker login -u <username> -p <password>`

#### Step 2: Create a docker file

Here we will use our github folder [**docker/build_fpga_server_docker**](https://github.com/Xilinx/FPGA_as_a_Service/tree/master/k8s-fpga-device-plugin/docker/build_fpga_server_docker)  as an example. In this folder, **"server"** is a file folder that is to be added into our docker image. It has four files:

|File | Description|
|---|---|
| [fpga_algo.awsxclbin](https://github.com/Xilinx/FPGA_as_a_Service/blob/master/k8s-fpga-device-plugin/docker/build_fpga_server_docker/server/fpga_algo.awsxclbin "fpga_algo.awsxclbin")| This is the xclbin of the algorithm implemented on FPGA.|
| [fpga_host_exe](https://github.com/Xilinx/FPGA_as_a_Service/blob/master/k8s-fpga-device-plugin/docker/build_fpga_server_docker/server/fpga_host_exe "fpga_host_exe") | This is the host executable that downloads the xclbin to FPGA and interacts with the FPGA. |
|  [fpga_server.py](https://github.com/Xilinx/FPGA_as_a_Service/blob/master/k8s-fpga-device-plugin/docker/build_fpga_server_docker/server/fpga_server.py "fpga_server.py")|This is a representative server program that calls the host executable and has ability to receive command from a client. One can merge this with host executable into one single server program in C++.|
|   [run.sh](https://github.com/Xilinx/FPGA_as_a_Service/blob/master/k8s-fpga-device-plugin/docker/build_fpga_server_docker/server/run.sh "run.sh")  |  This sets environment and calls  | fpga_server.py.

You can add any number of folders with any contents you need for your server to work.

Here we use **xilinxatg/aws-fpga-verify:20200131**  on docker hub as the base image mentioned earlier. In this example, the folder  **server**  will be added to an example location **/opt/xilinx/k8s/** in the docker image.

`#touch Dockerfile  `

create a dockerfile under the same folder with server

`#vi Dockerfile  `

To add following two lines into Dockerfile

```
FROM xilinxatg/aws-fpga-verify:20200131  
COPY docker /opt/xilinx/k8s/server
```

You can also use a Ubuntu or Centos/Redhat docker image as base images, and write your own dockerfile to build a docker image.
For example:

```
#example dockerfile
FROM ubuntu:18.04  #use ubuntu18.04 as base image
RUN apt-get update; apt-get install -y zip sudo git python; mkdir /tmp/deploy   #install needed packages
COPY u30_ubuntu_1804_v1.0_20201215.zip /tmp/deploy/                             #copy needed packages(this can be your xrt package) into image
RUN unzip /tmp/deploy/u30_ubuntu_1804_v1.0_20201215.zip -d /tmp/deploy/
WORKDIR /tmp/deploy/u30_ubuntu_1804_v1.0_20201215
RUN ./install.sh                                                                #install packages
```

#### Step 3: Build new docker image

`#docker build -t xilinxatg/k8s-plugin-dev:accelator_pod .  `

It will build a new docker image called  **accelerator_pod**  using the docker file "**Dockerfile**" under the current folder

`#docker images  `

You can run this command to check whether the new images  **accelerator_pod**  was created.

`# docker run -it <imageID>`  

To test the docker image you just created, run the above. You should see the folder  **server** added into the docker image.

#### Step 4: Push new image into docker hub

`#docker push xilinxatg/k8s-plugin-dev:accelator_pod`




#### Step 5: Create a docker image for client

Please repeat the steps 2 to 4 with your desired executable contents for the client to create another docker image called  **test_client_pod**. You can use [build_test_client_docker](https://github.com/Xilinx/FPGA_as_a_Service/tree/master/k8s-fpga-device-plugin/docker/build_test_client_docker)  as an example.



### 7.3 Verify docker image

Use the yaml files: [aws-accelator-pod.yaml](https://github.com/Xilinx/FPGA_as_a_Service/blob/master/k8s-fpga-device-plugin/aws-accelator-pod.yaml)  and [aws-test-client-pod.yaml](https://github.com/Xilinx/FPGA_as_a_Service/blob/master/k8s-fpga-device-plugin/aws-test-client-pod.yaml)  to create accelerator and client pods respectively.


#### Step 1: Create accelerator and client pods

`#kubectl create -f aws-accelator-pod.yaml  `

`#kubectl create -f aws-test-client-pod.yaml`

#### Step 2: Check pod status

After creating the two pods, there will be an accelerator pod with FPGA access, a client pod without FPGA access, an accelerator pod deployment service and a fpga-server-svc network service as shown below.

`#kubectl get pod`

```
NAME                          READY      STATUS    RESTARTS    AGE  
accelator-pod-ff67ff8b8-mwff   1/1       Running       0       22h  
test-client-pod                1/1       Running       0       23h
```

`#kubectl get deployment`

```
NAME             READY   UP-TO-DATE   AVAILABLE    AGE  
accelator-pod     1/1        1            1        22h
```

`#kubectl get service`

```
NAME               TYPE       CLUSTER-IP     EXTERNAL-IP    PORT(S)       AGE  
fpga-server-svc  NodePort     10.96.59.3        <none>   8010:31600/TCP   22h  
kubernetes       ClusterIP     10.96.0.1        <none>      443/TCP       14d
```

#### Step 3: Run hello world in client pod

`#kubectl exec test-client-pod python /opt/xilinx/k8s/client/client.py`



**Note:**

**If the status of the accelerator pod shows as pending, please check whether the card is already assigned to another running pod. If so, please delete the running pod and recreate the accelerator pod.**

