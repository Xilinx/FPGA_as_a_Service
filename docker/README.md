## How to build new docker image and test it in k8s-fpga-device-plugin

We will use an example to explain how a new docker image with desired contents such as your xclbin, your host code etc. can be built. Please note that any accelerator (FPGA) docker image should be derived form the base docker Xilinx image **xilinxatg/aws-fpga-verify:20200131** already hosted at the Docker Hub.

### Prerequisites

To host a docker image, you need some sort of service. You can host it locally if you like (please read online docker instructions for that). However, this instruction uses [Docker Hub](https://hub.docker.com/)  as the hosting service.

Go to [https://hub.docker.com/signup](https://hub.docker.com/signup) to create a Docker Hub account (if you do not have one already), and then create a docker repository.

In this document, we are using an example docker account named as **memo40k**  and an example repository named as **k8s.** Please substitute these by your account and repository names respectively. Also, please note that you can set your repository as private if you do not want others to see it.

### Prepare docker images

####   
Step 1: Login to your Docker Hub account

`#docker login -u <username> -p <password>`

#### Step 2: Create a docker file

Here we will use our github folder [**docker/build_fpga_server_docker**](https://github.com/Xilinx/FPGA_as_a_Service/tree/master/k8s-fpga-device-plugin/trunk/docker/build_fpga_server_docker)  as an example. In this folder, **"server"** is a file folder that is to be added into our docker image. It has four files:

|File | Description|
|---|---|
| [fpga_algo.awsxclbin](https://github.com/Xilinx/FPGA_as_a_Service/blob/master/k8s-fpga-device-plugin/trunk/docker/build_fpga_server_docker/server/fpga_algo.awsxclbin "fpga_algo.awsxclbin")| This is the xclbin of the algorithm implemented on FPGA.|
| [fpga_host_exe](https://github.com/Xilinx/FPGA_as_a_Service/blob/master/k8s-fpga-device-plugin/trunk/docker/build_fpga_server_docker/server/fpga_host_exe "fpga_host_exe") | This is the host executable that downloads the xclbin to FPGA and interacts with the FPGA. |
|  [fpga_server.py](https://github.com/Xilinx/FPGA_as_a_Service/blob/master/k8s-fpga-device-plugin/trunk/docker/build_fpga_server_docker/server/fpga_server.py "fpga_server.py")|This is a representative server program that calls the host executable and has ability to receive command from a client. One can merge this with host executable into one single server program in C++.|
|   [run.sh](https://github.com/Xilinx/FPGA_as_a_Service/blob/master/k8s-fpga-device-plugin/trunk/docker/build_fpga_server_docker/server/run.sh "run.sh")  |  This sets environment and calls  | fpga_server.py.

You can add any number of folders with any contents you need for your server to work.

The **xilinxatg/aws-fpga-verify:20200131**  on docker hub is the base image as mentioned earlier. In this example, the folder  **server**  will be added to an example location **/opt/xilinx/k8s/** in the docker image.

`#touch Dockerfile  `
create a dockerfile under the same folder with server

`#vi Dockerfile  `
To add following two lines into Dockerfile
```
FROM xilinxatg/aws-fpga-verify:20200131  
COPY docker /opt/xilinx/k8s/server
```
#### Step 3: Build new docker image

`#docker build -t memo40k/k8s:accelator_pod .  `
It will build a new docker image called  **accelerator_pod**  using the docker file "**Dockerfile**" under the current folder
`#docker images  `
You can run this command to check whether the new images  **accelerator_pod**  was created.
`# docker run -it <imageID>`  
To test the docker image you just created, run the above. You should see the folder  **server** added into the docker image.

#### Step 4: Push new image into docker hub

`#docker push memo40k/k8s:accelator_podk8`

You are all set.



#### Step 5: Create a docker image for client

Please repeat the steps 2 to 4 with your desired executable contents for the client to create another docker image called  **test_client_pod**. You can use [build_test_client_docker](https://github.com/Xilinx/FPGA_as_a_Service/tree/master/k8s-fpga-device-plugin/trunk/docker/build_test_client_docker)  as an example.



### Verify docker image

Use the yaml files: [aws-accelator-pod.yaml](https://github.com/Xilinx/FPGA_as_a_Service/blob/master/k8s-fpga-device-plugin/trunk/aws-accelator-pod.yaml)  and [aws-test-client-pod.yaml](https://github.com/Xilinx/FPGA_as_a_Service/blob/master/k8s-fpga-device-plugin/trunk/aws-test-client-pod.yaml)  to create accelerator and client pods respectively.



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

