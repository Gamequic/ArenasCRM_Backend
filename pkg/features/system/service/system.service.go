package systemservice

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/Gamequic/LivePreviewBackend/utils"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
)

var Logger = utils.NewLogger()

func GetSystemMetrics(prevNetStats *[]net.IOCountersStat, prevDiskStats *map[string]disk.IOCountersStat) (map[string]interface{}, error) {
	// Obtener detalles del sistema operativo
	var osDetails string
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("cmd", "/C", "ver")
		output, err := cmd.Output()
		if err != nil {
			osDetails = "No se pudo obtener la versión de Windows"
		} else {
			osDetails = strings.TrimSpace(string(output))
		}
	case "linux":
		cmd := exec.Command("cat", "/etc/os-release")
		output, err := cmd.Output()
		if err != nil {
			osDetails = "No se pudo obtener la distribución de Linux"
		} else {
			for _, line := range strings.Split(string(output), "\n") {
				if strings.HasPrefix(line, "PRETTY_NAME") {
					osDetails = strings.Trim(strings.Split(line, "=")[1], `"`)
					break
				}
			}
		}
	default:
		osDetails = "Sistema operativo no soportado para más detalles"
	}

	cpuPercent, err := cpu.Percent(0, false)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo CPU: %v", err)
	}

	memStats, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo memoria: %v", err)
	}

	netStats, err := net.IOCounters(true)
	if err != nil {
		Logger.Error(fmt.Sprintf("Error obteniendo estadísticas de red: %v", err))
		netStats = []net.IOCountersStat{}
	}

	diskPartitions, err := disk.Partitions(false)
	if err != nil {
		Logger.Error(fmt.Sprintf("Error obteniendo particiones de disco: %v", err))
		diskPartitions = []disk.PartitionStat{}
	}

	// Calcular velocidades de red por interfaz
	interfaces := []map[string]interface{}{}
	if len(*prevNetStats) > 0 {
		for i, iface := range netStats {
			if i >= len(*prevNetStats) {
				continue
			}
			prevIface := (*prevNetStats)[i]
			interfaces = append(interfaces, map[string]interface{}{
				"name":       iface.Name,
				"bytesSent":  iface.BytesSent,
				"bytesRecv":  iface.BytesRecv,
				"readSpeed":  float64(iface.BytesRecv-prevIface.BytesRecv) / 1024.0,
				"writeSpeed": float64(iface.BytesSent-prevIface.BytesSent) / 1024.0,
			})
		}
	}
	*prevNetStats = netStats

	// Calcular velocidades de discos
	diskSpeeds := map[string]interface{}{}
	diskIOStats, err := disk.IOCounters()
	if err != nil {
		diskSpeeds["isAvailable"] = false
	} else {
		for diskName, ioStat := range diskIOStats {
			readSpeed := 0.0
			writeSpeed := 0.0

			if prevDiskStats != nil {
				if prevStats, ok := (*prevDiskStats)[diskName]; ok {
					readSpeed = float64(ioStat.ReadBytes-prevStats.ReadBytes) / 1024.0
					writeSpeed = float64(ioStat.WriteBytes-prevStats.WriteBytes) / 1024.0
				}
			}

			diskSpeeds[diskName] = map[string]float64{
				"readSpeed":  readSpeed,
				"writeSpeed": writeSpeed,
			}
		}
		*prevDiskStats = diskIOStats
	}

	// Crear lista de particiones con espacio disponible
	diskDetails := []map[string]interface{}{}
	for _, partition := range diskPartitions {
		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			Logger.Error(fmt.Sprintf("Error obteniendo uso de la partición %s: %v", partition.Mountpoint, err))
			continue
		}
		diskDetails = append(diskDetails, map[string]interface{}{
			"mountpoint": partition.Mountpoint,
			"total":      usage.Total / 1024 / 1024,
			"free":       usage.Free / 1024 / 1024,
			"used":       usage.Used / 1024 / 1024,
		})
	}

	return map[string]interface{}{
		"os": map[string]interface{}{
			"name":    runtime.GOOS,
			"details": osDetails,
		},
		"cpuUsage":    cpuPercent[0],
		"totalMemory": memStats.Total / 1024 / 1024,
		"memoryUsage": memStats.Used / 1024 / 1024,
		"network": map[string]interface{}{
			"interfaces": interfaces,
		},
		"diskDetails": diskDetails,
		"diskSpeeds":  diskSpeeds,
	}, nil
}
