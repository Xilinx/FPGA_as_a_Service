# xilinx fpga device plugin on AWS F1
## Note

* XRT 2019.2+ is required. Older version of XRT doesn't work
* The awsxclbin used by the helloworld may not be accessible in all AWS regions. So far it is only
  accessible in the follow regions,
  -	us-east-1 (N.Virginia) 
  -	us-west-2 (Oregon)
  -	eu-west-1 (Ireland)
  -	asia-pacific (Sydeny)

Please run the following cmd to see if it is available in your F1.
```
# fpga-load-local-image -S 0 -I agfi-08025ce1d75d038c0
AFI          0       agfi-08025ce1d75d038c0  loaded            0        ok               0       0x04261818
AFIDEVICE    0       0x1d0f      0xf010      0000:00:1d.0
```

## Build XRT on F1

```
git clone http://github.com/aws/aws-fgpa.git
git clone -b 2019.2.0.3 http://github.com/xilinx/xrt.git
xrt/src/runtime_src/tools/scripts/xrtdeps.sh
scl enable devtoolset-6 bash  (centos only, ubuntu skip)
cd aws-fpga
source sdaccel_setup.sh
cd ../xrt/build
./build.sh
```

## Install XRT

```
cd xrt/build/Release
yum install ./name_of_xrt_pkg
yum install ./name_of_aws_pkg
```

or

```
cd xrt/build/Release
apt install ./name_of_xrt_pkg
apt install ./name_of_aws_pkg
```

check mpd status and fpga status

```
# systemctl status mpd
● mpd.service - Xilinx Management Proxy Daemon (MPD)
   Loaded: loaded (/etc/systemd/system/mpd.service; enabled; vendor preset: disabled)
   Active: active (running) since Thu 2020-02-06 23:21:10 UTC; 3 days ago
 Main PID: 10978 (mpd)
    Tasks: 3
   Memory: 20.0K
   CGroup: /system.slice/mpd.service
           └─10978 /opt/xilinx/xrt/bin/mpd
# /opt/xilinx/xrt/bin/xbutil scan
INFO: Found total 1 card(s), 1 are usable
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
System Configuration
OS name:	Linux
Release:	3.10.0-1062.4.1.el7.x86_64
Version:	#1 SMP Fri Oct 18 17:15:30 UTC 2019
Machine:	x86_64
Glibc:		2.17
Distribution:	CentOS Linux 7 (Core)
Now:		Mon Feb 10 18:12:21 2020
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
XRT Information
Version:	2.3.0
Git Hash:	9e13d57c4563e2c19bf5f518993f6e5a8dadc18a
Git Branch:	HEAD
Build Date:	2020-02-06 23:07:50
XOCL:		2.3.0,9e13d57c4563e2c19bf5f518993f6e5a8dadc18a
XCLMGMT:	2.3.0,9e13d57c4563e2c19bf5f518993f6e5a8dadc18a
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
 [0] 0000:00:1d.0 xilinx_aws-vu9p-f1_dynamic_5_0(ts=0xabcd) user(inst=128)

```

## Install docker

```
yum install docker
systemctl enable docker
```

or

```
apt install docker.io
systemctl enable docker
```

# Install kubernetes
please refer,

https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/
https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/create-cluster-kubeadm/
	
```
swapoff -a
cat <<EOF > /etc/yum.repos.d/kubernetes.repo
[kubernetes]
name=Kubernetes
baseurl=https://packages.cloud.google.com/yum/repos/kubernetes-el7-x86_64
enabled=1
gpgcheck=1
repo_gpgcheck=1
gpgkey=https://packages.cloud.google.com/yum/doc/yum-key.gpg https://packages.cloud.google.com/yum/doc/rpm-package-key.gpg
EOF
setenforce 0
sed -i 's/^SELINUX=enforcing$/SELINUX=permissive/' /etc/selinux/config
yum install -y kubelet kubeadm kubectl --disableexcludes=kubernetes
systemctl enable --now kubelet
```

or

```
swapoff -a
sudo apt-get update && sudo apt-get install -y apt-transport-https curhttps://kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/lhttps://kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/
curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add -
cat <<EOF | sudo tee /etc/apt/sources.list.d/kubernetes.list
deb https://apt.kubernetes.io/ kubernetes-xenial main
EOF
sudo apt-get update
sudo apt-get install -y kubelet kubeadm kubectl
sudo apt-mark hold kubelet kubeadm kubectl
```

## Create k8s cluster

```
kubeadm init
export KUBECONFIG=/etc/kubernetes/admin.conf
kubectl apply -f https://docs.projectcalico.org/v3.11/manifests/calico.yaml
kubectl taint nodes --all node-role.kubernetes.io/master-
```

## Deploy FPGA device plugin & helloworld add

```
kubectl create -f fpga-device-plugin.yaml
kubectl create -f aws-verify.yaml

```

## Run helloworld

check pod status, there will be a server pod with FPGA access and a client pod without FPGA access

```
# kubectl get pod
NAME                           READY   STATUS    RESTARTS   AGE
test-server-76759df476-xgv6x   1/1     Running   0          84m
testpod                        1/1     Running   0          84m
```

run helloworld -- client pod sends request to server pod, and server pod run helloworld on FPGA, then sends the output back to client as response

```
# kubectl exec testpod python /opt/test/server-test.py client
Send request to server...
Response from server:
Found Platform
Platform Name: Xilinx
Found Device=xilinx_aws-vu9p-f1_dynamic_5_0
INFO: Reading /opt/test/vector_addition_hw.awsxclbin
Loading: '/opt/test/vector_addition_hw.awsxclbin'
Result = 
Hello World !!! 
Hello World !!! 
Hello World !!! 
Hello World !!! 
Hello World !!! 
Hello World !!! 
Hello World !!! 
Hello World !!! 
Hello World !!! 
Hello World !!! 
Hello World !!! 
Hello World !!! 
Hello World !!! 
Hello World !!! 
Hello World !!! 
Hello World !!! 
TEST PASSED

--END--
```

## Author

Brian Xu(brianx@xilinx.com)
