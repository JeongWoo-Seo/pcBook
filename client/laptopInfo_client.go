package client

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/JeongWoo-Seo/pcBook/pb"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
)

func StartSenderWorker(laptopClient *LaptopClient, queue <-chan *pb.LaptopInfo) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream, err := laptopClient.service.SendLaptopInfo(ctx)
	if err != nil {
		log.Printf("Failed to open stream: %v", err)
		return
	}

	log.Println("Sender worker started...")

	for info := range queue {
		req := &pb.SendLaptopInfoRequest{
			Laptop: info,
		}

		err := stream.Send(req)
		if err != nil {
			log.Printf("Error while sending to stream: %v", err)
			break
		}
		log.Printf("Sent laptop info: %s", info.GetId())
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Printf("Error while receiving final response: %v", err)
		return
	}

	log.Printf("Final Server Response: %s", res.GetMsg())
	log.Println("Sender worker finished successfully.")
}

func (laptopClient *LaptopClient) GetBatteryInfo() (uint32, error) {
	cmd := exec.Command("pmset", "-g", "batt")
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("error running pmset command: %w", err)
	}

	outputStr := string(output)
	lines := strings.Split(outputStr, "\n")

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		if strings.Contains(trimmedLine, "%") {

			// '%' 이전의 문자열만 남깁니다.
			parts := strings.Split(trimmedLine, "%")
			if len(parts) < 1 {
				continue
			}

			stringBeforePercent := parts[0] // 예: " -InternalBattery-0 (id=4194403) 29"

			// " -InternalBattery-0","(id=4194403)","29"
			fields := strings.Fields(stringBeforePercent)
			if len(fields) < 1 {
				continue
			}

			// 3. 잔량 숫자만 추출하고 공백 제거
			percentageStr := strings.TrimSpace(fields[len(fields)-1])

			// 4. 숫자로 변환
			percentageInt, err := strconv.Atoi(percentageStr)
			if err != nil {
				return 0, fmt.Errorf("failed to parse battery percentage '%s': %w", percentageStr, err)
			}

			return uint32(percentageInt), nil
		}
	}

	return 0, fmt.Errorf("battery infor not found")
}

func (laptopClient *LaptopClient) GetCPUInfo() (float64, error) {
	// 1초 동안의 전체 CPU 평균 사용률을 측정
	percents, err := cpu.Percent(time.Second, false)
	if err != nil {
		return 0, fmt.Errorf("error getting CPU percent: %w", err)
	}

	if len(percents) < 1 {
		return 0, fmt.Errorf("cpu percentage result is empty")

	}

	return percents[0], nil
}

func (laptopClient *LaptopClient) GetMemoryInfo() (*pb.MemoryUsage, *pb.MemoryUsage, error) {
	v, err := mem.VirtualMemory()
	if err != nil {
		return nil, nil, fmt.Errorf("error getting RAM info: %w", err)
	}

	const GB float64 = 1024 * 1024 * 1024
	ramTotalGB := float64(v.Total) / GB
	ramUsedGB := float64(v.Used) / GB
	ramUsagePercent := v.UsedPercent

	ramUsageInfo := &pb.MemoryUsage{
		TotalMemory:   ramTotalGB,
		CurrentMemory: ramUsedGB,
		Usage:         ramUsagePercent,
	}

	// macOS의 경우, 루트 디렉토리("/")
	u, err := disk.Usage("/")
	if err != nil {
		return nil, nil, fmt.Errorf("error getting disk info: %w", err)
	}

	diskTotalGB := float64(u.Total) / GB
	diskUsedGB := float64(u.Used) / GB

	storageUsage := &pb.MemoryUsage{
		TotalMemory:   diskTotalGB,
		CurrentMemory: diskUsedGB,
		Usage:         u.UsedPercent,
	}

	return ramUsageInfo, storageUsage, nil
}

func (laptopClient *LaptopClient) GetNetInfo() (*pb.Network, error) {
	startStats, err := net.IOCounters(true)
	if err != nil {
		return nil, fmt.Errorf("error getting net start stats: %w", err)
	}

	time.Sleep(1 * time.Second)

	endStats, err := net.IOCounters(true)
	if err != nil {
		return nil, fmt.Errorf("error getting net end stats: %w", err)
	}

	var totalRxDelta uint64 = 0
	var totalTxDelta uint64 = 0

	//  모든 인터페이스의 변화량을 합산 wifi,eth0,eth1
	if len(startStats) != len(endStats) {
		return nil, fmt.Errorf("net interface count changed during measurement")
	}

	for i := range startStats {
		start := startStats[i]
		end := endStats[i]

		totalRxDelta += end.BytesRecv - start.BytesRecv
		totalTxDelta += end.BytesSent - start.BytesSent
	}

	result := &pb.Network{
		Rx: totalRxDelta,
		Tx: totalTxDelta,
	}
	return result, nil
}

func (laptopClient *LaptopClient) GetMacSerialID() (string, error) {
	cmd := exec.Command("system_profiler", "SPHardwareDataType")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("system_profiler 명령어 실행 오류: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// 시리얼 번호가 포함된 라인 확인
		if strings.HasPrefix(trimmedLine, "Serial Number (system):") {
			// 콜론과 공백을 기준으로 문자열을 분리합니다.
			parts := strings.Split(trimmedLine, ":")
			if len(parts) > 1 {
				// 마지막 부분의 앞뒤 공백을 제거하여 순수한 시리얼 번호를 반환합니다.
				serial := strings.TrimSpace(parts[1])
				return serial, nil
			}
		}
	}

	// 시리얼 번호를 찾지 못한 경우
	return "", fmt.Errorf("시리얼 번호 정보를 찾을 수 없습니다")
}
