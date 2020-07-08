package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
)

// Stats Object to be turned into JSON and posted to the API
type Stats struct {
	SysInfo struct {
		Hostname string
		Platform string
		CPU      struct {
			CPU        int32   `json:"cpu"`
			VendorID   string  `json:"vendorId"`
			Family     string  `json:"family"`
			Model      string  `json:"model"`
			PhysicalID string  `json:"physicalId"`
			CoreID     string  `json:"coreId"`
			Cores      int32   `json:"cores"`
			ModelName  string  `json:"modelName"`
			Mhz        float64 `json:"mhz"`
			CacheSize  int32   `json:"cacheSize"`
		}
		RAM  uint64
		Disk uint64
	}
	MemUsage struct {
		Total       uint64
		Free        uint64
		UsedPercent float64
	}
	LoadAverage struct {
		Min1  float64
		Min5  float64
		Min15 float64
	}
	CPUUsage []float64
	Uptime   string
}

func main() {
	APIUrl := "http://localhost:8000"

	fmt.Printf("Starting Monitor...\n")
	getURL := fmt.Sprintf("%s/api/get/%s", APIUrl, "1234")
	response, err := http.Get(getURL)
	if err != nil {
		fmt.Println("ERR => Request failed with error $s\n", err)
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		fmt.Println(string(data))
	}

	hostStat, _ := host.Info()
	cpuStat, _ := cpu.Info()
	vmStat, _ := mem.VirtualMemory()
	diskStat, _ := disk.Usage("\\")

	info := new(Stats)

	info.SysInfo.Hostname = hostStat.Hostname
	info.SysInfo.Platform = hostStat.Platform
	info.SysInfo.CPU.CPU = cpuStat[0].CPU
	info.SysInfo.CPU.VendorID = cpuStat[0].VendorID
	info.SysInfo.CPU.Model = cpuStat[0].Model
	info.SysInfo.CPU.PhysicalID = cpuStat[0].PhysicalID
	info.SysInfo.CPU.CoreID = cpuStat[0].CoreID
	info.SysInfo.CPU.Cores = cpuStat[0].Cores
	info.SysInfo.CPU.ModelName = cpuStat[0].ModelName
	info.SysInfo.CPU.Mhz = cpuStat[0].Mhz
	info.SysInfo.CPU.CacheSize = cpuStat[0].CacheSize
	info.SysInfo.RAM = vmStat.Total / 1024 / 1024
	info.SysInfo.Disk = diskStat.Total / 1024 / 1024

	info.MemUsage.Total = vmStat.Total
	info.MemUsage.Free = vmStat.Total
	info.MemUsage.UsedPercent = vmStat.UsedPercent

	load, _ := load.Avg()
	info.LoadAverage.Min1 = load.Load1
	info.LoadAverage.Min5 = load.Load5
	info.LoadAverage.Min15 = load.Load15

	percent, _ := cpu.Percent(time.Second, true)
	fmt.Printf("  CPU: %.2f\n", percent)
	info.CPUUsage = percent

	uptime, _ := host.Uptime()
	days := uptime / (60 * 60 * 24)
	hours := (uptime - (days * 60 * 60 * 24)) / (60 * 60)
	minutes := ((uptime - (days * 60 * 60 * 24)) - (hours * 60 * 60)) / 60
	info.Uptime = fmt.Sprintf("%d days, %d hours, %d minutes", days, hours, minutes)

	jsonValue, _ := json.Marshal(info)

	//jsonData := map[string]struct{"sys": info, "memory": memValue}
	postURL := fmt.Sprintf("%s/api/data", APIUrl)
	response, err = http.Post(postURL, "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		fmt.Println("ERR => Request failed with error $s\n", err)
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		fmt.Println(string(data))
	}
	fmt.Println("Monitor Completed.")
}
