// Copyright 2018 Xilinx Corporation. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"io/ioutil"
	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1beta1"
	"os"
	"path"
	"strconv"
	"strings"
)

const (
	SysfsDevices   = "/sys/bus/pci/devices"
	MgmtPrefix     = "/dev/xclmgmt"
	UserPrefix     = "/dev/dri"
	UserPFKeyword  = "drm"
	DRMSTR         = "renderD"
	DSAverFile     = "VBNV"
	DSAtsFile      = "timestamp"
	DSAinfoFile    = "rom.u."
	InstanceFile   = "instance"
	MgmtFunc       = ".1"
	UserFunc       = ".0"
	MgmtFile       = "mgmt_pf"
	UserFile       = "user_pf"
	VendorFile     = "vendor"
	DeviceFile     = "device"
	XilinxVendorID = "0x10ee"
)

type Pairs struct {
	Mgmt string
	User string
}

type Device struct {
	index     string
	shellVer  string
	timestamp string
	DBDF      string // this is for user pf
	deviceID  string //devid of the user pf
	Healthy   string
	Nodes     Pairs
}

func GetInstance(DBDF string) (string, error) {
	strArray := strings.Split(DBDF, ":")
	domain, err := strconv.ParseUint(strArray[0], 16, 16)
	if err != nil {
		return "", fmt.Errorf("strconv failed: %s\n", strArray[0])
	}
	bus, err := strconv.ParseUint(strArray[1], 16, 8)
	if err != nil {
		return "", fmt.Errorf("strconv failed: %s\n", strArray[1])
	}
	strArray = strings.Split(strArray[2], ".")
	dev, err := strconv.ParseUint(strArray[0], 16, 8)
	if err != nil {
		return "", fmt.Errorf("strconv failed: %s\n", strArray[0])
	}
	fc, err := strconv.ParseUint(strArray[1], 16, 8)
	if err != nil {
		return "", fmt.Errorf("strconv failed: %s\n", strArray[1])
	}
	ret := domain*65536 + bus*256 + dev*8 + fc
	return strconv.FormatUint(ret, 10), nil
}

func GetUserPF(dir string) (string, error) {
	userFiles, err := ioutil.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("Can't read folder %s \n", dir)
	}
	for _, userFile := range userFiles {
		fname := userFile.Name()

		if !strings.HasPrefix(fname, DRMSTR) {
			continue
		}
		return fname, nil
	}
	return "", nil
}

func GetFileContent(file string) (string, error) {
	if buf, err := ioutil.ReadFile(file); err != nil {
		return "", fmt.Errorf("Can't read file %s \n", file)
	} else {
		ret := strings.Trim(string(buf), "\n")
		return ret, nil
	}
}

//Prior to 2018.3 release, Xilinx FPGA has mgmt PF as func 1 and user PF
//as func 0. The func numbers of the 2 PFs are swapped after 2018.3 release.
//The FPGA device driver in (and after) 2018.3 release creates sysfs file --
//mgmt_pf and user_pf accordingly to reflect what a PF really is.
//the k8s fpga plugin may manage a cluster with hybrid FPGAs (with old and new
//DSAs), so we need to have a workaround here to distinguish user and mgmt PFs.
//the logic is as follows:
//if (pci func is 1) { //it may be mgmt func of old dsa or user func of new dsa
//    if (there is no user_pf file) { // old dsa
//        this is mgmt func
//    }
//} else { // it may be user func of old dsa or mgmt func of new dsa
//    if (there exists mgmt_pf file) { // new dsa
//        this is mgmt func
//    }
//}
//TODO: In the future, an API should be introduced in xrt so that the plugin
//does not need to know the hardware and low level software changes.
//But even with an API, we still need to handle old DSA & XRT. So the workaround
//is still necessary. Sigh...
func FileExist(fname string) bool {
	if _, err := os.Stat(fname); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func IsMgmtPf(pciID string) bool {
	if strings.HasSuffix(pciID, MgmtFunc) {
		fname := path.Join(SysfsDevices, pciID, UserFile)
		if !FileExist(fname) {
			return true
		}
		return false
	} else {
		fname := path.Join(SysfsDevices, pciID, MgmtFile)
		if FileExist(fname) {
			return true
		}
		return false
	}
}

func GetDevices() ([]Device, error) {
	var devices []Device
	pairMap := make(map[string]*Pairs)
	pciFiles, err := ioutil.ReadDir(SysfsDevices)
	if err != nil {
		return nil, fmt.Errorf("Can't read folder %s \n", SysfsDevices)
	}

	for _, pciFile := range pciFiles {
		pciID := pciFile.Name()

		fname := path.Join(SysfsDevices, pciID, VendorFile)
		vendorID, err := GetFileContent(fname)
		if err != nil {
			return nil, err
		}
		if strings.EqualFold(vendorID, XilinxVendorID) != true {
			continue
		}

		DBD := pciID[:len(pciID)-2]
		if _, ok := pairMap[DBD]; !ok {
			pairMap[DBD] = &Pairs{
				Mgmt: "",
				User: "",
			}
		}

		// For containers deployed on top of baremetal machines, xilinx FPGA
		// in sysfs will always appear as pair of mgmt PF and user PF
		// For containers deployed on top of VM, there may be only user PF
		// available(mgmt PF is not assigned to the VM)
		// so mgmt in Pair may be empty
		if !IsMgmtPf(pciID) { //user pf
			userDBDF := pciID
			instance, err := GetInstance(userDBDF)
			if err != nil {
				return nil, err
			}
			// get dsa version
			fname = path.Join(SysfsDevices, pciID, DSAinfoFile+instance, DSAverFile)
			content, err := GetFileContent(fname)
			if err != nil {
				return nil, err
			}
			dsaVer := content
			// get dsa timestamp
			fname = path.Join(SysfsDevices, pciID, DSAinfoFile+instance, DSAtsFile)
			content, err = GetFileContent(fname)
			if err != nil {
				return nil, err
			}
			dsaTs := content
			// get device id
			fname = path.Join(SysfsDevices, pciID, DeviceFile)
			content, err = GetFileContent(fname)
			if err != nil {
				return nil, err
			}
			devid := content
			// get user PF node
			userpf, err := GetUserPF(path.Join(SysfsDevices, pciID, UserPFKeyword))
			if err != nil {
				return nil, err
			}
			userNode := path.Join(UserPrefix, userpf)
			pairMap[DBD].User = userNode

			//TODO: check temp, power, fan speed etc, to give a healthy level
			//so far, return Healthy
			healthy := pluginapi.Healthy
			devices = append(devices, Device{
				index:     strconv.Itoa(len(devices) + 1),
				shellVer:  dsaVer,
				timestamp: dsaTs,
				DBDF:      userDBDF,
				deviceID:  devid,
				Healthy:   healthy,
				Nodes:     *pairMap[DBD],
			})
		} else { //mgmt pf
			// get mgmt instance
			fname = path.Join(SysfsDevices, pciID, InstanceFile)
			content, err := GetFileContent(fname)
			if err != nil {
				return nil, err
			}
			pairMap[DBD].Mgmt = MgmtPrefix + content
		}

	}
	return devices, nil
}

/*
func main() {
	devices, err := GetDevices()
	if err != nil {
		fmt.Printf("%s !!!\n", err)
		return
	}
	for _, device := range devices {
		fmt.Printf("%v", device)
	}
}
*/
