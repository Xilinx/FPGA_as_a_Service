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
	"flag"
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"os"
	"strconv"
	"strings"
	"syscall"
)

var (
	U30NameConvention   = "CommonName"
	U30AllocUnit        = "Card"
	DeviceNameCustomize = "False"
	VirtualDev          = "False"
	VirtualNum          = 1
)

func main() {
	// Parse command-line arguments
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flagLogLevel := flag.String("log-level", "info", "Define the logging level: error, info, debug.")
	flag.Parse()

	switch *flagLogLevel {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	}

	version_path := "/opt/xilinx/k8s-device-plugin/version_num"
	if version_file, err := ioutil.ReadFile(version_path); err != nil {
		log.Println("Can't read version file %s", version_path)
	} else {
		version := strings.Trim(string(version_file), "\n")
		log.Println("Plugin Version:", version)
	}

	ReadNameType := os.Getenv("U30NameConvention")
	if strings.EqualFold(ReadNameType, "ExactName") {
		log.Println("Set U30NameConvention: ExactName")
		U30NameConvention = "ExactName"
	} else {
		log.Println("Set U30NameConvention: CommonName")
		U30NameConvention = "CommonName"
	}
	ReadAllocUnitType := os.Getenv("U30AllocUnit")
	if strings.EqualFold(ReadAllocUnitType, "Device") {
		log.Println("Set U30AllocUnit: Device")
		U30AllocUnit = "Device"
	} else {
		log.Println("Set U30AllocUnit: Card")
		U30AllocUnit = "Card"
	}
	ReadDeviceNameCustomize := os.Getenv("DeviceNameCustomize")
	if strings.EqualFold(ReadDeviceNameCustomize, "True") {
		log.Println("Set DeviceNameCustomize: True")
		DeviceNameCustomize = "True"
	} else {
		log.Println("Set DeviceNameCustomize: False")
		DeviceNameCustomize = "False"
	}

	ReadVirtualDev := os.Getenv("VirtualDev")
	if strings.EqualFold(ReadVirtualDev, "True") {
		log.Println("Virtual Device Mode: On")
		VirtualDev = "True"
	} else {
		log.Println("Virtual Device Mode: OFF")
		VirtualDev = "False"
	}
	ReadVirtualNum := os.Getenv("VirtualNum")
	VirtualNum, _ = strconv.Atoi(ReadVirtualNum)
	if VirtualNum < 1 {
		log.Warn("Invalid input for VirtualNum, will set VirtualNum as 1")
		VirtualNum = 1
	}
	log.Println("VirtualNum:", VirtualNum)

	log.Println("Starting FS watcher.")
	watcher, err := newFSWatcher(pluginapi.DevicePluginPath)
	if err != nil {
		log.Printf("Failed to created FS watcher: %s.", err)
		os.Exit(1)
	}
	defer watcher.Close()

	log.Println("Starting OS watcher.")
	sigs := newOSWatcher(syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	restart := true
	var devicePlugin *FPGADevicePlugin
L:
	for {
		if restart {
			devicePlugin = NewFPGADevicePlugin()
			restart = false
		}

		select {
		case update := <-devicePlugin.updateChan:
			devicePlugin.checkDeviceUpdate(update)

		case event := <-watcher.Events:
			if event.Name == pluginapi.KubeletSocket && event.Op&fsnotify.Create == fsnotify.Create {
				log.Printf("inotify: %s created, restarting.", pluginapi.KubeletSocket)
				restart = true
			}

		case err := <-watcher.Errors:
			log.Printf("inotify: %s", err)

		case s := <-sigs:
			switch s {
			case syscall.SIGHUP:
				log.Println("Received SIGHUP, restarting.")
				restart = true
			default:
				log.Printf("Received signal \"%v\", shutting down.", s)
				break L
			}
		}
	}
}
