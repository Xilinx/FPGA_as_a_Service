// Copyright 2018-2022, Xilinx, Inc.
// Copyright 2023, Advanced Micro Device, Inc.
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
	"bufio"
	"fmt"
	"io/ioutil"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
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
	SNSTRV70       = "hwmon_sdm.u."
	DSAverFile     = "VBNV"
	DSAtsFile      = "timestamp"
	InstanceFile   = "instance"
	UUID           = "logic_uuids"
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

const (
	DevicesPath       = "/dev"
	AmaDevicePrefix   = "ama_transcoder"
	MiscClassPath     = "/sys/class/misc"
	AmaBusId          = "bus_id"
	AmaDeivceInfo     = "device_info"
	AmaPlatformPrefix = "MA"
)

type Pairs struct {
	Mgmt string
	User string
	Qdma string
}

type Device struct {
	index      string
	shellVer   string
	deviceType string
	uuid       string
	timestamp  string
	DBDF       string // this is for user pf
	deviceID   string //devid of the user pf
	Healthy    string
	SN         string
	Nodes      *Pairs
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

func GetAlveoDevices() ([]Device, error) {
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
			// get dsa type from dsa version
			dsaType := strings.Split(dsaVer, "_")[1]
			// get dsa timestamp
			fname = path.Join(SysfsDevices, pciID, romFolder, DSAtsFile)
			content, err = GetFileContent(fname)
			if err != nil {
				return nil, err
			}
			dsaTs := content
			// get dsa uuid
			fname = path.Join(SysfsDevices, pciID, UUID)
			content, err = GetFileContent(fname)
			if err != nil {
				return nil, err
			}
			dsaUUID := content[len(content)-6 : len(content)]
			// get device id
			fname = path.Join(SysfsDevices, pciID, DeviceFile)
			content, err = GetFileContent(fname)
			if err != nil {
				return nil, err
			}
			devid := content

			//get file path for Serial Number
			SNFolder := ""
			if strings.EqualFold(dsaType, "v70") == true {
				SNFolder, err = GetFileNameFromPrefix(path.Join(SysfsDevices, pciID), SNSTRV70)
				if err != nil {
					return nil, err
				}

			} else {
				SNFolder, err = GetFileNameFromPrefix(path.Join(SysfsDevices, pciID), SNSTR)
				if err != nil {
					return nil, err
				}

			}
			// get Serial Number
			// AWS F1 device has no serial numbers, adding default serial number "F1-Node" for each AWS F1 device
			fname = path.Join(SysfsDevices, pciID, SNFolder, SNFile)
			content, err = GetFileContent(fname)
			if err != nil {
				if strings.EqualFold(vendorID, AWS_ID) == true {
					content = "F1-Node"
				} else if strings.EqualFold(dsaType, "u30") == true {
					fmt.Println("No Serial Number detected, Serial Number is must required for u30 device")
					return nil, err
				} else {
					fmt.Println("Device has no serial number detected")
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
			if strings.EqualFold(VirtualDev, "True") {
				for i := 0; i < VirtualNum; i++ {
					devices = append(devices, Device{
						index:      strconv.Itoa(len(devices) + 1),
						shellVer:   dsaVer,
						deviceType: dsaType,
						uuid:       dsaUUID,
						timestamp:  dsaTs,
						DBDF:       userDBDF + "-" + strconv.Itoa(i),
						deviceID:   devid,
						Healthy:    healthy,
						SN:         SN,
						Nodes:      pairMap[DBD],
					})
				}
			} else {
				devices = append(devices, Device{
					index:      strconv.Itoa(len(devices) + 1),
					shellVer:   dsaVer,
					deviceType: dsaType,
					uuid:       dsaUUID,
					timestamp:  dsaTs,
					DBDF:       userDBDF,
					deviceID:   devid,
					Healthy:    healthy,
					SN:         SN,
					Nodes:      pairMap[DBD],
				})
			}
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

// List all AMA devices
func GetAMADevices() ([]Device, error) {
	var devices []Device
	pairMap := make(map[string]*Pairs)
	if _, err := os.Stat(DevicesPath); os.IsNotExist(err) {
		// no devices path found
		return nil, err
	}
	devFiles, err := ioutil.ReadDir(DevicesPath)
	if err != nil {
		fmt.Errorf("Cannot read folder %s", DevicesPath)
	}

	for _, devFile := range devFiles {
		if devFile.IsDir() {
			// not a device
			continue
		}

		//renderID
		devId := devFile.Name()
		if !strings.HasPrefix(devId, AmaDevicePrefix) {
			// not an AMA device
			continue
		}

		//DBDF
		busId, err := GetFileContent(path.Join(MiscClassPath, devId, AmaBusId))
		if err != nil {
			return nil, err
		}

		DBD := busId[:len(busId)-2]
		if _, ok := pairMap[DBD]; !ok {
			pairMap[DBD] = &Pairs{
				Mgmt: "",
				User: "",
				Qdma: "",
			}
		}
		pairMap[DBD].User = path.Join(DevicesPath, devId)
		if err != nil {
			return nil, err
		}

		productNameKey, productSNKey, deviceIdKey := "Product name", "Product serial number", "PCIe device ID"
		// open additional AMA device info file
		file, err := os.Open(path.Join(MiscClassPath, devId, AmaDeivceInfo))
		defer file.Close()
		if err != nil {
			fmt.Errorf("Failed to open file path %s", path.Join(MiscClassPath, devId, AmaDeivceInfo))
			return nil, err
		}

		// read file line by line
		fscanner := bufio.NewScanner(file)
		fscanner.Split(bufio.ScanLines)
		boardName := ""
		SN := ""
		devid := ""
		for fscanner.Scan() {
			line := fscanner.Text()
			strs := strings.Split(line, "=")
			if len(strs) != 2 {
				continue
			}

			if strings.EqualFold(productNameKey, strings.TrimSpace(strs[0])) {
				//fmt.Printf("BoardName: %v \n", strings.TrimSpace(strs[1]))
				boardName = "MA35"

			}
			if strings.EqualFold(productSNKey, strings.TrimSpace(strs[0])) {
				//fmt.Printf("SerialNumber: %v \n", strings.TrimSpace(strs[1]))
				SN = strings.TrimSpace(strs[1])
			}
			if strings.EqualFold(deviceIdKey, strings.TrimSpace(strs[0])) {
				//fmt.Printf("DeviceID: %v \n", strings.TrimSpace(strs[1]))
				devid = strings.TrimSpace(strs[1])
			}
		}
		//TODO: check temp, power, fan speed etc, to give a healthy level
		//so far, return Healthy
		healthy := pluginapi.Healthy
		//healthy := "temp-health"
		if strings.EqualFold(VirtualDev, "True") {
			for i := 0; i < VirtualNum; i++ {
				devices = append(devices, Device{
					index:      strconv.Itoa(len(devices) + 1),
					shellVer:   boardName,
					deviceType: boardName,
					uuid:       "ma35",
					timestamp:  "0",
					DBDF:       busId + "-" + strconv.Itoa(i),
					deviceID:   devid,
					Healthy:    healthy,
					SN:         SN,
					Nodes:      pairMap[DBD],
				})
			}
		} else {
			devices = append(devices, Device{
				index:      strconv.Itoa(len(devices) + 1),
				shellVer:   boardName,
				deviceType: boardName,
				uuid:       "ma35",
				timestamp:  "0",
				DBDF:       busId,
				deviceID:   devid,
				Healthy:    healthy,
				SN:         SN,
				Nodes:      pairMap[DBD],
			})
		}
	}
	return devices, nil
}

func GetDevices() ([]Device, error) {
	AMADevicesArry, err := GetAMADevices()
	if err != nil {
		return nil, err
	}
	AlveoDevicesArry, err := GetAlveoDevices()
	if err != nil {
		return nil, err
	}
	//combine Alveo device list and AMA device list into one list
	for _, AlveoDevice := range AlveoDevicesArry {
		AMADevicesArry = append(AMADevicesArry, AlveoDevice)
	}
	return AMADevicesArry, err
}

/*
func main() {

	//ama_devices, err := GetAMADevices()
	//if err != nil {
	//	fmt.Printf("%s !!!\n", err)
	//	return
	//}
		ama_devices, err := GetAMADevices()
		if err != nil {
			fmt.Printf("%s !!!\n", err)
			return
		}
		for _, device := range ama_devices {
			fmt.Printf("AMADevice: %v \n", device)
			fmt.Printf("AMADriver: %v \n",device.Nodes)
		}
	all_devices, err := GetAMADevices()
	if err != nil {
		fmt.Printf("%s !!!\n", err)
		return
	}
	for _, device := range all_devices {
		fmt.Printf("AMA ShellVer: %v \n", device.shellVer)
		fmt.Printf("AMA Timestamp: %v \n", device.timestamp)
		fmt.Printf("AMA DBDF: %v \n", device.DBDF)
		fmt.Printf("AMA deviceID: %v \n", device.deviceID)
		fmt.Printf("AMA SN: %v \n", device.SN)
		fmt.Printf("AMA Driver: %v \n", device.Nodes)
	}
}
*/
