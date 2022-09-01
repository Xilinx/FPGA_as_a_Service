// Copyright 2018-2022 Xilinx Corporation. All Rights Reserved.
// Author: Brian Xu(brianx@xilinx.com)
// For technical support, please contact k8s_dev@amd.com
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
	"time"
)

const (
	SysfsDevices   = "/sys/bus/pci/devices"
	MgmtPrefix     = "/dev/xclmgmt"
	UserPrefix     = "/dev/dri"
	QdmaPrefix     = "/dev/xfpga"
	QDMASTR        = "dma.qdma.u"
	UserPFKeyword  = "drm"
	DRMSTR         = "renderD"
	ROMSTR         = "rom"
	SNSTR          = "xmc.u."
	DSAverFile     = "VBNV"
	DSAtsFile      = "timestamp"
	InstanceFile   = "instance"
	MgmtFile       = "mgmt_pf"
	UserFile       = "user_pf"
	VendorFile     = "vendor"
	DeviceFile     = "device"
	SNFile         = "serial_num"
	VtShell        = "xilinx_u30"
	U30CommonShell = "ama_u30"
	XilinxVendorID = "0x10ee"
	ADVANTECH_ID   = "0x13fe"
	AWS_ID         = "0x1d0f"
	AristaVendorID = "0x3475"
)

type Pairs struct {
	Mgmt string
	User string
	Qdma string
}

type Device struct {
	index     string
	shellVer  string
	timestamp string
	DBDF      string // this is for user pf
	deviceID  string //devid of the user pf
	Healthy   string
	SN        string
	Nodes     *Pairs
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
			strings.EqualFold(vendorID, AristaVendorID) != true &&
			strings.EqualFold(vendorID, AWS_ID) != true &&
			strings.EqualFold(vendorID, ADVANTECH_ID) != true {
			continue
		}

		DBD := pciID[:len(pciID)-2]
		if _, ok := pairMap[DBD]; !ok {
			pairMap[DBD] = &Pairs{
				Mgmt: "",
				User: "",
				Qdma: "",
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
			count := 0
			if err != nil {
				return nil, err
			}
			for romFolder == "" {
				if count >= 36 {
					break
				}
				time.Sleep(10 * time.Second)
				romFolder, err = GetFileNameFromPrefix(path.Join(SysfsDevices, pciID), ROMSTR)
				if romFolder != "" {
					time.Sleep(20 * time.Second)
					break
				}
				fmt.Println(count, pciID, romFolder, err)
				count += 1
			}
			SNFolder, err := GetFileNameFromPrefix(path.Join(SysfsDevices, pciID), SNSTR)
			if err != nil {
				return nil, err
			}
			// get dsa version
			fname = path.Join(SysfsDevices, pciID, romFolder, DSAverFile)
			content, err := GetFileContent(fname)
			if err != nil {
				return nil, err
			}
			if strings.EqualFold(U30NameConvention, "CommonName") && strings.Contains(content, VtShell) {
				content = U30CommonShell
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
			// get Serial Number
			// AWS F1 device has no serial numbers, adding default serial number "F1-Node" for each AWS F1 device
			fname = path.Join(SysfsDevices, pciID, SNFolder, SNFile)
			content, err = GetFileContent(fname)
			if err != nil {
				if strings.EqualFold(vendorID, AWS_ID) == true {
					content = "F1-Node"
				} else {
					return nil, err
				}
			}
			SN := content
			// get user PF node
			userpf, err := GetFileNameFromPrefix(path.Join(SysfsDevices, pciID, UserPFKeyword), DRMSTR)
			if err != nil {
				return nil, err
			}
			userNode := path.Join(UserPrefix, userpf)
			pairMap[DBD].User = userNode

			//get qdma device node if it exists
			instance, err := GetInstance(userDBDF)
			if err != nil {
				return nil, err
			}

			qdmaFolder, err := GetFileNameFromPrefix(path.Join(SysfsDevices, pciID), QDMASTR)
			if err != nil {
				return nil, err
			}

			if qdmaFolder != "" {
				pairMap[DBD].Qdma = path.Join(QdmaPrefix, QDMASTR+instance+".0")
			}

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
				SN:        SN,
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

/*
func main() {
        devices, err := GetDevices()
	if err != nil {
                fmt.Printf("%s !!!\n", err)
                return
        }

        //SNFolder, err := GetFileNameFromPrefix(path.Join(SysfsDevices, "0000:e3:00.1"), SNSTR)
	//fname := path.Join(SysfsDevices, "0000:e3:00.1", SNFolder, SNFile)
	//content, err := GetFileContent(fname)
	//SN := content
	//fmt.Printf("SN: %v \n", SN)
        for _, device := range devices {
                fmt.Printf("Device: %v \n", device)
                fmt.Printf("Timestamp: %v \n",device.timestamp)
                fmt.Printf("SN: %v  \n", device.SN)
                fmt.Printf("ID: %s  \n\n", device.deviceID)
        }
}
*/
