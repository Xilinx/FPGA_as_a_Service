// Copyright 2018 Xilinx Corporation. All Rights Reserved.
// Author: Brian Xu(brianx@xilinx.com)
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
	"regexp"
	"strconv"
	"strings"
)

const (
	SysfsDevices   = "/sys/bus/pci/devices"
	DTBDeviceModel = "/proc/device-tree/model"
	MgmtPrefix     = "/dev/xclmgmt"
	UserPrefix     = "/dev/dri"
	UserPFKeyword  = "drm"
	DRMSTR         = "renderD"
	ROMSTR         = "rom"
	DSAverFile     = "VBNV"
	DSAtsFile      = "timestamp"
	InstanceFile   = "instance"
	MgmtFunc       = ".1"
	UserFunc       = ".0"
	MgmtFile       = "mgmt_pf"
	UserFile       = "user_pf"
	VendorFile     = "vendor"
	DeviceFile     = "device"
	XilinxVendorID = "0x10ee"
	ADVANTECH_ID   = "0x13fe"
	AWS_ID         = "0x1d0f"
	//Zynq Supported boards: https://xilinx.github.io/XRT/2019.2/html/platforms.html
	//DTB model name pattern: https://github.com/Xilinx/linux-xlnx/tree/master/arch/arm64/boot/dts/xilinx
	ZynqRegex      = "ZynqMP"
	ZynqModelRegex = "ZCU1(9|0[246])"
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
	Nodes     *Pairs
}

func GetFileNameFromPrefix(dir string, prefix string) (string, error) {
	userFiles, err := ioutil.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("Can't read folder %s", dir)
	}
	for _, userFile := range userFiles {
		fname := userFile.Name()

		if !strings.HasPrefix(fname, prefix) {
			continue
		}
		return fname, nil
	}
	return "", nil
}

func GetFileContent(file string) (string, error) {
	if buf, err := ioutil.ReadFile(file); err != nil {
		return "", fmt.Errorf("Can't read file %s", file)
	} else {
		ret := strings.Trim(string(buf), "\n")
		return ret, nil
	}
}

//Prior to 2018.3 release, Xilinx FPGA has mgmt PF as func 1 and user PF
//as func 0. The func numbers of the 2 PFs are swapped after 2018.3 release.
//The FPGA device driver in (and after) 2018.3 release creates sysfs file --
//mgmt_pf and user_pf accordingly to reflect what a PF really is.
//
//The plugin will rely on this info to determine whether the a entry is mgmtPF,
//userPF, or none. This also means, it will not support 2018.2 any more.
func FileExist(fname string) bool {
	if _, err := os.Stat(fname); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func IsMgmtPf(pciID string) bool {
	fname := path.Join(SysfsDevices, pciID, MgmtFile)
	return FileExist(fname)
}

func IsUserPf(pciID string) bool {
	fname := path.Join(SysfsDevices, pciID, UserFile)
	return FileExist(fname)
}

func GetDevices() ([]Device, error) {
	//Check if we are on a Zynq board.
	modelName, err := GetFileContent(DTBDeviceModel)
	if err == nil {
		if strings.Contains(modelName, ZynqRegex) {
			return GetDevicesEdge(modelName)
		}
	}
	var devices []Device
	pairMap := make(map[string]*Pairs)
	pciFiles, err := ioutil.ReadDir(SysfsDevices)
	if err != nil {
		return nil, fmt.Errorf("Can't read folder %s", SysfsDevices)
	}

	for _, pciFile := range pciFiles {
		pciID := pciFile.Name()

		fname := path.Join(SysfsDevices, pciID, VendorFile)
		vendorID, err := GetFileContent(fname)
		if err != nil {
			return nil, err
		}
		if strings.EqualFold(vendorID, XilinxVendorID) != true &&
			strings.EqualFold(vendorID, AWS_ID) != true &&
			strings.EqualFold(vendorID, ADVANTECH_ID) != true {
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
		if IsUserPf(pciID) { //user pf
			userDBDF := pciID
			romFolder, err := GetFileNameFromPrefix(path.Join(SysfsDevices, pciID), ROMSTR)
			if err != nil {
				return nil, err
			}
			// get dsa version
			fname = path.Join(SysfsDevices, pciID, romFolder, DSAverFile)
			content, err := GetFileContent(fname)
			if err != nil {
				return nil, err
			}
			dsaVer := content
			// get dsa timestamp
			fname = path.Join(SysfsDevices, pciID, romFolder, DSAtsFile)
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
			userpf, err := GetFileNameFromPrefix(path.Join(SysfsDevices, pciID, UserPFKeyword), DRMSTR)
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
				Nodes:     pairMap[DBD],
			})
		} else if IsMgmtPf(pciID) { //mgmt pf
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
func GetDevicesEdge(modelName string) ([]Device, error) {
	var devices []Device
	devid := strconv.Itoa(len(devices) + 1)
	zynqRegex := regexp.MustCompile(ZynqRegex)
	modelRegex := regexp.MustCompile(ZynqModelRegex)
	board := zynqRegex.FindString(modelName)
	model := modelRegex.FindString(modelName)
	pairMap := make(map[string]*Pairs)
	pairMap[devid] = &Pairs{
		Mgmt: "",
		User: "",
	}
	//TODO: check temp, power, fan speed etc, to give a healthy level
	//so far, return Healthy
	healthy := pluginapi.Healthy
	// get DRM render node
	drm, err := GetFileNameFromPrefix(UserPrefix, DRMSTR)
	if err != nil {
		return nil, err
	}
	userNode := path.Join(UserPrefix, drm)
	pairMap[devid].User = userNode

	//TODO: check for more devices in edge boards, as for now
	//in XRT number of devices is hardcoded to 1. See:
	//XRT/src/runtime_src/core/edge/tools/xbutil/xbutil.cpp:402
	//For shellver and timestamp we are using the board and model
	//to be able to orchestrate on different boards.
	devices = append(devices, Device{
		index:     strconv.Itoa(len(devices) + 1),
		DBDF:      devid,
		deviceID:  devid,
		shellVer:  board,
		timestamp: model,
		Healthy:   healthy,
		Nodes:     pairMap[devid],
	})
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
