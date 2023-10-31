#/bin/bash
#
# Copyright 2020-2022, Xilinx, Inc.
# Copyright 2023, Advanced Micro Device, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# Make sure this script is sourced
script=${BASH_SOURCE[0]}
if [ $script == $0 ]; then
    echo "ERROR: You must source this script"
    exit 2
fi

source /opt/xilinx/xrt/setup.sh

#get all AMD-Xilinx UserPF devices BDF
accelerator_device_pcie=()
getBDF(){
	Check_device=`lspci | grep "Xilinx"`
	array=(${Check_device//"Processing accelerators: Xilinx Corporation Device"/})
	for(( i=0; i<${#array[@]}; i++ )) 
	do
		if [[ ${array[i]} =~ [0-9a-z][0-9a-z]:* ]]; then
			if [ -s /sys/bus/pci/devices/"0000:"${array[i]}/user_pf ]; then
				accelerator_device_pcie+=("0000:"${array[i]})
			fi
		fi
	done
}


echo ""
echo "All AMD-Xilinx Devices:"
echo ""
getBDF  #get all xilinx Processing accelerators device BDF
cnt=0   #card_number

#print accelerator device
for pcie in ${accelerator_device_pcie[@]}
do
                card_value=""
                shellVer=`cat /sys/bus/pci/devices/$pcie/rom.*/VBNV`
                IFS='_' read -r -a array <<< "$shellVer"

                echo device $cnt :
		printf "%-10s %-10s\n" Keyword Value
                printf "%-10s %-10s\n" ======================
                printf "%-10s %-10s\n" deviceType ${array[1]}
                printf "%-10s %-10s\n" DBDF $pcie
                printf "%-10s %-10s\n" shellVer `cat /sys/bus/pci/devices/$pcie/rom.*/VBNV`
                printf "%-10s %-10s\n" uuid `cat /sys/bus/pci/devices/$pcie/logic_uuids`
                if [[ ${array[1]} == "v70" ]]; then
                        printf "%-10s %-10s\n" SN `cat /sys/bus/pci/devices/$pcie/hwmon_sdm.u.*/serial_num`
                else
                        printf "%-10s %-10s\n" SN `cat /sys/bus/pci/devices/$pcie/xmc.u.*/serial_num`
                fi
                printf "%-10s %-10s\n" timestamp `cat /sys/bus/pci/devices/$pcie/rom.*/timestamp`
                echo ""
                cnt=$(($cnt+1))
done


#get MA35 device ID
media_device_name=()
getAMAdevices(){
        Check_device=`ls /dev | grep "ama_transcoder"`
	array=(${Check_device//" "/})
        for(( i=0; i<${#array[@]}; i++ ))
        do
                media_device_name+=(${array[i]})
        done
}

getAMAdevices

#print multimedia controller device
for device in ${media_device_name[@]}
do

                echo device $cnt :
                printf "%-10s %-10s\n" Keyword Value
                printf "%-10s %-10s\n" ======================
                printf "%-10s %-10s\n" deviceType MA35
                printf "%-10s %-10s\n" DBDF `cat /sys/class/misc/$device/bus_id`
                printf "%-10s %-10s\n" shellVer MA35
                printf "%-10s %-10s\n" uuid ma35
                printf "%-10s %-10s\n" SN `cat /sys/class/misc/$device/device_info | grep -oP '(?<=Product serial number = )\w+'`
                printf "%-10s %-10s\n" timestamp 0
                echo ""
                cnt=$(($cnt+1))
done

