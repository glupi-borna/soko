package system

import (
	"time"
	"net/http"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/net"
)

var memVars = map[string]any{
	"VirtualMemory": mem.VirtualMemory,
	"VirtualMemoryEx": mem.VirtualMemoryEx,
	"SwapMemory": mem.SwapMemory,
}

var cpuVars = map[string]any{
	"Percent": cpu.Percent,
	"Info": cpu.Info,
	"Counts": cpu.Counts,
	"Times": cpu.Times,
}

var diskVars = map[string]any {
	"Partitions": disk.Partitions,
	"Usage": disk.Usage,
	"SerialNumber": disk.SerialNumber,
	"IOCounters": disk.IOCounters,
}

const secondsToNanoSeconds = 1000000000

var (
	netCounterFirst = true

	bytesSent uint64 = 0
	bytesSentPrev uint64 = 0

	bytesRecv uint64 = 0
	bytesRecvPrev uint64 = 0

	netCounterUp uint64 = 0
	netCounterDown uint64 = 0
	netCounterLatency time.Duration

	lastPingTime time.Time
	pingInterval time.Duration = time.Duration(10 * secondsToNanoSeconds)
	pingAddress = "http://1.1.1.1"
)

func netCountersUpdate(no_ping bool) {
	counters, err := net.IOCounters(false)
	if err != nil { return }
	if len(counters) < 1 { return }
	counter := counters[0]

	bytesRecvPrev = bytesRecv
	bytesSentPrev = bytesSent

	bytesRecv = counter.BytesRecv
	bytesSent = counter.BytesSent

	if netCounterFirst == true {
		netCounterFirst = false
		bytesRecvPrev = bytesRecv
		bytesRecvPrev = bytesSent
	}

	netCounterUp = bytesSent - bytesSentPrev
	netCounterDown = bytesRecv - bytesRecvPrev

	if !no_ping {
		now := time.Now()
		if now.Sub(lastPingTime) > pingInterval {
			go netPing()
			lastPingTime = now
		}
	}
}

func netPing() {
	start_time := time.Now()
	_, err := http.Get(pingAddress)
	if err != nil { println(err.Error()); return }
	netCounterLatency = time.Now().Sub(start_time)
}

func UpDownLatency(no_ping bool, do_update bool) (uint64, uint64, time.Duration) {
	if do_update { netCountersUpdate(no_ping) }
	return netCounterUp, netCounterDown, netCounterLatency
}

func SetPingIntervalSeconds(seconds float64) {
	ns := uint64(seconds * float64(secondsToNanoSeconds))
	pingInterval = time.Duration(ns)
}

func SetPingAddress(addr string) {
	pingAddress = addr
}

var netVars = map[string]any {
	"Connections": net.Connections,
	"ConnectionsPid": net.ConnectionsPid,
	"IOCounters": net.IOCounters,
	"Interfaces": net.Interfaces,
	"ProtoCounters": net.ProtoCounters,
	"SetPingIntervalSeconds": SetPingIntervalSeconds,
	"SetPingAddress": SetPingAddress,
	"UpDownLatency": UpDownLatency,
}

var WidgetVars = map[string]any{
	"MEM": memVars,
	"CPU": cpuVars,
	"DISK": diskVars,
	"NET": netVars,
}
