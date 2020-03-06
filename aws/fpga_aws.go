// Copyright 2019 Xilinx Corporation. All Rights Reserved.
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
	"path"
	"strconv"
	"strings"
)

const (
	SysfsDevices           = "/sys/bus/pci/devices"
	UserPrefix             = "/dev/dri"
	UserPFKeyword          = "drm"
	DRMSTR                 = "renderD"
	VendorFile             = "vendor"
	DeviceFile             = "device"
	AWSF1VendorID          = "0x1d0f"
	AWSF1UserPFDeviceID    = "0x1042"
	AWSF1UserPFDeviceIDSdx = "0xf010"
	AWSDSAVer              = "xilinx_aws-vu9p-f1-04261818_dynamic_5_0" //hard coded so far
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
		return "", fmt.Errorf("Can't read folder %s \n", dir)
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
		return "", fmt.Errorf("Can't read file %s \n", file)
	} else {
		ret := strings.Trim(string(buf), "\n")
		return ret, nil
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
		if strings.EqualFold(vendorID, AWSF1VendorID) != true {
			continue
		}
		fname = path.Join(SysfsDevices, pciID, DeviceFile)
		devID, err := GetFileContent(fname)
		if err != nil {
			return nil, err
		}
		if strings.EqualFold(devID, AWSF1UserPFDeviceID) != true &&
			strings.EqualFold(devID, AWSF1UserPFDeviceIDSdx) != true {
			continue
		}

		DBD := pciID[:len(pciID)-2]
		if _, ok := pairMap[DBD]; !ok {
			pairMap[DBD] = &Pairs{
				Mgmt: "",
				User: "",
			}
		}

		userDBDF := pciID
		// get dsa version. So far the info is hard coded and not populated into sysfs.
		dsaVer := AWSDSAVer
		// get dsa timestamp. So far the info is not used by aws. using "0" here
		dsaTs := "0"
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
			deviceID:  devID,
			Healthy:   healthy,
			Nodes:     pairMap[DBD],
		})
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
