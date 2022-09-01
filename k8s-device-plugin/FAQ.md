# FAQ  
## Question: When I reserve a machine looking for a U250, I am assigned a node with a U250, but from within the pod, I can see the U250 and also the U280 (which is also present on that node). Is this intended behavior? ##

**Answer:** This issue has been fixed form kubernetes version 1.17, if you are using previous versions of kubernetes you can try updating your kubernetes cluster to the latest version. The previous problem is multiple types of FGPA cards or Shell on one node can not be handled by Kubernetes. You can Check the following link for detailed info: [https://github.com/kubernetes/kubernetes/issues/70350](https://github.com/kubernetes/kubernetes/issues/70350) 

## Question: When testing Vitis in AWS FPGA environment. There is an AFI agfi-069ddd533a748059b which is first loaded when I do the systemctl start mpd for the very first time when I boot the machine. Then, at the end inside my-pod when running the ./helloworld vector_addition_hw.awsxclbin, this time is the one associated to the vector_addition_hw.awsxclbin, AFI agfi-2 is loaded, and the one effectively used. Is both have “vector_addition_hw” in their name? ##

 **Answer:** AWS F1 allows a user to use FPGA in two ways:  traditional hardware design flow using [HDK](https://clicktime.symantec.com/3DiMRHsPYvzA8YqaAYokmLV6H2?u=https%3A%2F%2Fgithub.com%2Faws%2Faws-fpga%2Fblob%2Fmaster%2Fhdk%2FREADME.md) and  2) software like flow using Use [SDAccel/Vitis](https://clicktime.symantec.com/3MGuf45R2JanX4SS6wJtNTM6H2?u=https%3A%2F%2Fgithub.com%2Faws%2Faws-fpga%2Fblob%2Fmaster%2FSDAccel%2FREADME.md). The flow you are using is the SDAccel/Vitis one. In order to differentiate the two flows, AWS came up with a device ID scheme that Xilinx adheres to. There is a Xilinx Run Time deamon named [MPD](https://clicktime.symantec.com/3Rz2catG2XLeoDpWiMmdeq36H2?u=https%3A%2F%2Fxilinx.github.io%2FXRT%2Fmaster%2Fhtml%2Fcloud_vendor_support.html) that is started using the command “systemctl start mpd”  as part of installing XRT. This daemon is required to download the prebuilt AFI “agfi-069ddd533a748059b” to allow AWS hardware to differentiate the SDAccel/Vitis from form the HDK flow. When user application runs, the AFI is replaced by the user AFI. Every time you reboot the machine, the [MPD](https://clicktime.symantec.com/3Rz2catG2XLeoDpWiMmdeq36H2?u=https%3A%2F%2Fxilinx.github.io%2FXRT%2Fmaster%2Fhtml%2Fcloud_vendor_support.html) will be restarted and hence you will see that agfi-069ddd533a748059b got installed. Once the Accelerator Pod runs, it could install the AFI of interest.

## Question: One application fails to run inside container, with possible error like “Failed to find Xilinx platform”. It can run well outside container. xbutil list show the device existing inside container, same as that outside container. ##

**Answer:** Linux is using ICDs ("Installable Client Drivers") to setup OpenCL. /etc/OpenCL/vendors/xilinx.icd is used to tell the ICD loader what OpenCL implementations (ICDs) are installed on the system. Some application is directly link to Linux system standard OpenCL lib but NOT Xilinx specified OpenCL lib. For this case, the OpenCL APIs in application will fail if the icd file is NOT set correctly.

The above problem is due to missing /etc/OpenCL/vendors/xilinx.icd inside container. Using following command to copy /etc/OpenCL/vendors/xilinx.icd (with one line “/opt/xilinx/xrt/lib/libxilinxopencl.so”) from host into container can solve this issue.
```
docker cp /etc/OpenCL/vendors/xilinx.icd containerID:/etc/OpenCL/vendors/xilinx.icd
```

## Question: K8s can not pulling image when creating pod on Redhat worker nodes with image from Private Docker Repo. ##

**Answer:**
If you want to pulling from a private dockerhub repository or using a private registry, there are three ways to do it, we have verified the solultoin 1 and 2 work on Redhat node. This situation only happens when you use secret to pull images on a Redhat Node.

##### 1) Configuring nodes to authenticate to a private registry
If you run Docker on your nodes, you can configure the Docker container runtime to authenticate to a private container registry. This approach is suitable if you can control node configuration. Docker stores keys for private registries in the $HOME/.dockercfg or $HOME/.docker/config.json file. If you put the same file in the search paths {cwd of kubelet}/config.json, kubelet uses it as the credential provider when pulling images.

Following command need to be run as root user on all nodes you need pulling images from a private repo:
```
docker login
cp /root/.docker/config.json /var/lib/kubelet/
```
##### 2) Pre-pulled images
By default, the kubelet tries to pull each image from the specified registry. However, if the imagePullPolicy property of the container is set to IfNotPresent or Never, then a local image is used (preferentially or exclusively, respectively).

##### 3) Specifying imagePullSecrets on a Pod
You can use following command to create a secret with a docker config:

```
kubectl create secret docker-registry REGISTRY_KEY_NAME \
  --docker-server=DOCKER_REGISTRY_SERVER \
  --docker-username=DOCKER_USER \
  --docker-password=DOCKER_PASSWORD \
  --docker-email=DOCKER_EMAIL
```

And now, you can create pods which reference that secret by adding an imagePullSecrets section to a Pod definition.
The secret may not work properly when trying to create a pod on redhat worker node, you can reference other solutions to temporary fix it.
```
#example yaml file for creating pod pulling from private repo with secret
apiVersion: v1
kind: Pod
metadata:
  name: foo
  namespace: awesomeapps
spec:
  containers:
    - name: foo
      image: janedoe/awesomeapp:v1
  imagePullSecrets:
    - name: REGISTRY_KEY_NAME
```

You can also check k8s document for more detialed information: https://kubernetes.io/docs/concepts/containers/images
