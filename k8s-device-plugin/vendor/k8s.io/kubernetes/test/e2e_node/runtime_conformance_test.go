/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package e2e_node

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/kubernetes/pkg/kubelet/images"
	"k8s.io/kubernetes/test/e2e/framework"
	"k8s.io/kubernetes/test/e2e_node/services"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gomegatypes "github.com/onsi/gomega/types"
)

const (
	consistentCheckTimeout = time.Second * 5
	retryTimeout           = time.Minute * 5
	pollInterval           = time.Second * 1
)

var _ = framework.KubeDescribe("Container Runtime Conformance Test", func() {
	f := framework.NewDefaultFramework("runtime-conformance")

	Describe("container runtime conformance blackbox test", func() {
		Context("when starting a container that exits", func() {
			framework.ConformanceIt("it should run with the expected status", func() {
				restartCountVolumeName := "restart-count"
				restartCountVolumePath := "/restart-count"
				testContainer := v1.Container{
					Image: busyboxImage,
					VolumeMounts: []v1.VolumeMount{
						{
							MountPath: restartCountVolumePath,
							Name:      restartCountVolumeName,
						},
					},
				}
				testVolumes := []v1.Volume{
					{
						Name: restartCountVolumeName,
						VolumeSource: v1.VolumeSource{
							EmptyDir: &v1.EmptyDirVolumeSource{Medium: v1.StorageMediumMemory},
						},
					},
				}
				testCases := []struct {
					Name          string
					RestartPolicy v1.RestartPolicy
					Phase         v1.PodPhase
					State         ContainerState
					RestartCount  int32
					Ready         bool
				}{
					{"terminate-cmd-rpa", v1.RestartPolicyAlways, v1.PodRunning, ContainerStateRunning, 2, true},
					{"terminate-cmd-rpof", v1.RestartPolicyOnFailure, v1.PodSucceeded, ContainerStateTerminated, 1, false},
					{"terminate-cmd-rpn", v1.RestartPolicyNever, v1.PodFailed, ContainerStateTerminated, 0, false},
				}
				for _, testCase := range testCases {

					// It failed at the 1st run, then succeeded at 2nd run, then run forever
					cmdScripts := `
f=%s
count=$(echo 'hello' >> $f ; wc -l $f | awk {'print $1'})
if [ $count -eq 1 ]; then
	exit 1
fi
if [ $count -eq 2 ]; then
	exit 0
fi
while true; do sleep 1; done
`
					tmpCmd := fmt.Sprintf(cmdScripts, path.Join(restartCountVolumePath, "restartCount"))
					testContainer.Name = testCase.Name
					testContainer.Command = []string{"sh", "-c", tmpCmd}
					terminateContainer := ConformanceContainer{
						PodClient:     f.PodClient(),
						Container:     testContainer,
						RestartPolicy: testCase.RestartPolicy,
						Volumes:       testVolumes,
						PodSecurityContext: &v1.PodSecurityContext{
							SELinuxOptions: &v1.SELinuxOptions{
								Level: "s0",
							},
						},
					}
					terminateContainer.Create()
					defer terminateContainer.Delete()

					By("it should get the expected 'RestartCount'")
					Eventually(func() (int32, error) {
						status, err := terminateContainer.GetStatus()
						return status.RestartCount, err
					}, retryTimeout, pollInterval).Should(Equal(testCase.RestartCount))

					By("it should get the expected 'Phase'")
					Eventually(terminateContainer.GetPhase, retryTimeout, pollInterval).Should(Equal(testCase.Phase))

					By("it should get the expected 'Ready' condition")
					Expect(terminateContainer.IsReady()).Should(Equal(testCase.Ready))

					status, err := terminateContainer.GetStatus()
					Expect(err).ShouldNot(HaveOccurred())

					By("it should get the expected 'State'")
					Expect(GetContainerState(status.State)).To(Equal(testCase.State))

					By("it should be possible to delete [Conformance]")
					Expect(terminateContainer.Delete()).To(Succeed())
					Eventually(terminateContainer.Present, retryTimeout, pollInterval).Should(BeFalse())
				}
			})

			rootUser := int64(0)
			nonRootUser := int64(10000)
			for _, testCase := range []struct {
				name      string
				container v1.Container
				phase     v1.PodPhase
				message   gomegatypes.GomegaMatcher
			}{
				{
					name: "if TerminationMessagePath is set [Conformance]",
					container: v1.Container{
						Image:   busyboxImage,
						Command: []string{"/bin/sh", "-c"},
						Args:    []string{"/bin/echo -n DONE > /dev/termination-log"},
						TerminationMessagePath: "/dev/termination-log",
						SecurityContext: &v1.SecurityContext{
							RunAsUser: &rootUser,
						},
					},
					phase:   v1.PodSucceeded,
					message: Equal("DONE"),
				},

				{
					name: "if TerminationMessagePath is set as non-root user and at a non-default path [Conformance]",
					container: v1.Container{
						Image:   busyboxImage,
						Command: []string{"/bin/sh", "-c"},
						Args:    []string{"/bin/echo -n DONE > /dev/termination-custom-log"},
						TerminationMessagePath: "/dev/termination-custom-log",
						SecurityContext: &v1.SecurityContext{
							RunAsUser: &nonRootUser,
						},
					},
					phase:   v1.PodSucceeded,
					message: Equal("DONE"),
				},

				{
					name: "from log output if TerminationMessagePolicy FallbackToLogOnError is set [Conformance]",
					container: v1.Container{
						Image:   busyboxImage,
						Command: []string{"/bin/sh", "-c"},
						Args:    []string{"/bin/echo -n DONE; /bin/false"},
						TerminationMessagePath:   "/dev/termination-log",
						TerminationMessagePolicy: v1.TerminationMessageFallbackToLogsOnError,
					},
					phase:   v1.PodFailed,
					message: Equal("DONE\n"),
				},

				{
					name: "as empty when pod succeeds and TerminationMessagePolicy FallbackToLogOnError is set",
					container: v1.Container{
						Image:   busyboxImage,
						Command: []string{"/bin/sh", "-c"},
						Args:    []string{"/bin/echo DONE; /bin/true"},
						TerminationMessagePath:   "/dev/termination-log",
						TerminationMessagePolicy: v1.TerminationMessageFallbackToLogsOnError,
					},
					phase:   v1.PodSucceeded,
					message: Equal(""),
				},

				{
					name: "from file when pod succeeds and TerminationMessagePolicy FallbackToLogOnError is set [Conformance]",
					container: v1.Container{
						Image:   busyboxImage,
						Command: []string{"/bin/sh", "-c"},
						Args:    []string{"/bin/echo -n OK > /dev/termination-log; /bin/echo DONE; /bin/true"},
						TerminationMessagePath:   "/dev/termination-log",
						TerminationMessagePolicy: v1.TerminationMessageFallbackToLogsOnError,
					},
					phase:   v1.PodSucceeded,
					message: Equal("OK"),
				},
			} {
				It(fmt.Sprintf("should report termination message %s", testCase.name), func() {
					testCase.container.Name = "termination-message-container"
					c := ConformanceContainer{
						PodClient:     f.PodClient(),
						Container:     testCase.container,
						RestartPolicy: v1.RestartPolicyNever,
					}

					By("create the container")
					c.Create()
					defer c.Delete()

					By(fmt.Sprintf("wait for the container to reach %s", testCase.phase))
					Eventually(c.GetPhase, retryTimeout, pollInterval).Should(Equal(testCase.phase))

					By("get the container status")
					status, err := c.GetStatus()
					Expect(err).NotTo(HaveOccurred())

					By("the container should be terminated")
					Expect(GetContainerState(status.State)).To(Equal(ContainerStateTerminated))

					By("the termination message should be set")
					Expect(status.State.Terminated.Message).Should(testCase.message)

					By("delete the container")
					Expect(c.Delete()).To(Succeed())
				})
			}
		})

		Context("when running a container with a new image", func() {
			// The service account only has pull permission
			auth := `
{
	"auths": {
		"https://gcr.io": {
			"auth": "Replace with the public auth code from k8s community",
			"email": "image-pulling@authenticated-image-pulling.iam.gserviceaccount.com"
		}
	}
}`
			secret := &v1.Secret{
				Data: map[string][]byte{v1.DockerConfigJsonKey: []byte(auth)},
				Type: v1.SecretTypeDockerConfigJson,
			}
			// The following images are not added into NodeImageWhiteList, because this test is
			// testing image pulling, these images don't need to be prepulled. The ImagePullPolicy
			// is v1.PullAlways, so it won't be blocked by framework image white list check.
			for _, testCase := range []struct {
				description        string
				image              string
				secret             bool
				credentialProvider bool
				phase              v1.PodPhase
				waiting            bool
			}{
				{
					description: "should not be able to pull image from invalid registry",
					image:       "invalid.com/invalid/alpine:3.1",
					phase:       v1.PodPending,
					waiting:     true,
				},
				{
					description: "should not be able to pull non-existing image from gcr.io",
					image:       "gcr.io/google_containers/invalid-image:invalid-tag",
					phase:       v1.PodPending,
					waiting:     true,
				},
				{
					description: "should be able to pull image from gcr.io",
					image:       "gcr.io/google_containers/alpine-with-bash:1.0",
					phase:       v1.PodRunning,
					waiting:     false,
				},
				{
					description: "should be able to pull image from docker hub",
					image:       "alpine:3.1",
					phase:       v1.PodRunning,
					waiting:     false,
				},
				{
					description: "should not be able to pull from private registry without secret",
					image:       "gcr.io/authenticated-image-pulling/alpine:3.1",
					phase:       v1.PodPending,
					waiting:     true,
				},
				{
					description: "should be able to pull from private registry with secret",
					image:       "gcr.io/authenticated-image-pulling/alpine:3.1",
					secret:      true,
					phase:       v1.PodRunning,
					waiting:     false,
				},
				{
					description:        "should be able to pull from private registry with credential provider",
					image:              "gcr.io/authenticated-image-pulling/alpine:3.1",
					credentialProvider: true,
					phase:              v1.PodRunning,
					waiting:            false,
				},
			} {
				testCase := testCase
				It(testCase.description+" [Conformance]", func() {
					name := "image-pull-test"
					command := []string{"/bin/sh", "-c", "while true; do sleep 1; done"}
					container := ConformanceContainer{
						PodClient: f.PodClient(),
						Container: v1.Container{
							Name:    name,
							Image:   testCase.image,
							Command: command,
							// PullAlways makes sure that the image will always be pulled even if it is present before the test.
							ImagePullPolicy: v1.PullAlways,
						},
						RestartPolicy: v1.RestartPolicyNever,
					}
					if testCase.secret {
						secret.Name = "image-pull-secret-" + string(uuid.NewUUID())
						By("create image pull secret")
						_, err := f.ClientSet.CoreV1().Secrets(f.Namespace.Name).Create(secret)
						Expect(err).NotTo(HaveOccurred())
						defer f.ClientSet.CoreV1().Secrets(f.Namespace.Name).Delete(secret.Name, nil)
						container.ImagePullSecrets = []string{secret.Name}
					}
					if testCase.credentialProvider {
						configFile := filepath.Join(services.KubeletRootDirectory, "config.json")
						err := ioutil.WriteFile(configFile, []byte(auth), 0644)
						Expect(err).NotTo(HaveOccurred())
						defer os.Remove(configFile)
					}
					// checkContainerStatus checks whether the container status matches expectation.
					checkContainerStatus := func() error {
						status, err := container.GetStatus()
						if err != nil {
							return fmt.Errorf("failed to get container status: %v", err)
						}
						// We need to check container state first. The default pod status is pending, If we check
						// pod phase first, and the expected pod phase is Pending, the container status may not
						// even show up when we check it.
						// Check container state
						if !testCase.waiting {
							if status.State.Running == nil {
								return fmt.Errorf("expected container state: Running, got: %q",
									GetContainerState(status.State))
							}
						}
						if testCase.waiting {
							if status.State.Waiting == nil {
								return fmt.Errorf("expected container state: Waiting, got: %q",
									GetContainerState(status.State))
							}
							reason := status.State.Waiting.Reason
							if reason != images.ErrImagePull.Error() &&
								reason != images.ErrImagePullBackOff.Error() {
								return fmt.Errorf("unexpected waiting reason: %q", reason)
							}
						}
						// Check pod phase
						phase, err := container.GetPhase()
						if err != nil {
							return fmt.Errorf("failed to get pod phase: %v", err)
						}
						if phase != testCase.phase {
							return fmt.Errorf("expected pod phase: %q, got: %q", testCase.phase, phase)
						}
						return nil
					}
					// The image registry is not stable, which sometimes causes the test to fail. Add retry mechanism to make this
					// less flaky.
					const flakeRetry = 3
					for i := 1; i <= flakeRetry; i++ {
						var err error
						By("create the container")
						container.Create()
						By("check the container status")
						for start := time.Now(); time.Since(start) < retryTimeout; time.Sleep(pollInterval) {
							if err = checkContainerStatus(); err == nil {
								break
							}
						}
						By("delete the container")
						container.Delete()
						if err == nil {
							break
						}
						if i < flakeRetry {
							framework.Logf("No.%d attempt failed: %v, retrying...", i, err)
						} else {
							framework.Failf("All %d attempts failed: %v", flakeRetry, err)
						}
					}
				})
			}
		})
	})
})
