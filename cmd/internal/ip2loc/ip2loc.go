package ip2loc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

// IPData 表示一条 IP 范围和对应的地理信息
type IPData struct {
	StartIP     uint32 // 起始 IP
	EndIP       uint32 // 结束 IP
	LocationIdx uint32 // 地理信息在数据区中的索引
	prefix      uint32 // 前缀（IP 的第一个八位字节）
}

// LocationData 表示去重后的地理信息
type LocationData struct {
	Offset uint32 // 在数据区中的偏移量
	Length uint32 // 地理信息的长度
	Text   string // 地理信息字符串
}

func Convert(inputFile, outputFile string) {
	ipDataList, locations, err := loadIPDataFromFile(inputFile)
	if err != nil {
		fmt.Println("加载数据失败:", err)
		return
	}
	err = generateIPDat(outputFile, ipDataList, locations)
	if err != nil {
		fmt.Println("生成文件失败:", err)
	} else {
		fmt.Printf("生成文件成功: %s\n", outputFile)
	}
}

// 从文本行解析 TXT 格式的 IPData 和 LocationData
func parseIPData(line string, locationMap map[string]uint32, locations *[]LocationData) (IPData, error) {
	fields := strings.SplitN(line, "|", 15)
	if len(fields) < 15 {
		for len(fields) < 15 {
			fields = append(fields, "")
		}
	}

	startIP := ipToUint32(fields[0])
	endIP := ipToUint32(fields[1])
	location := strings.Join(fields[4:15], "|")

	var locIdx uint32
	if offset, exists := locationMap[location]; exists {
		locIdx = offset
	} else {
		locIdx = uint32(len(*locations))
		locationMap[location] = locIdx
		*locations = append(*locations, LocationData{Text: location})
	}

	return IPData{
		StartIP:     startIP,
		EndIP:       endIP,
		LocationIdx: locIdx,
		prefix:      startIP >> 24,
	}, nil
}

// 从文本行解析 CSV 格式的 IPData 和 LocationData
func parseCSVData(line string, locationMap map[string]uint32, locations *[]LocationData) (IPData, error) {
	fields := strings.Split(line, ",")
	if len(fields) < 15 {
		for len(fields) < 15 {
			fields = append(fields, "")
		}
	}

	// 移除引号并填充到 15 个字段
	for i, field := range fields {
		fields[i] = strings.Trim(field, `"`)
	}

	startIP := ipToUint32(fields[0])
	endIP := ipToUint32(fields[1])
	location := strings.Join(fields[4:15], "|")

	var locIdx uint32
	if offset, exists := locationMap[location]; exists {
		locIdx = offset
	} else {
		locIdx = uint32(len(*locations))
		locationMap[location] = locIdx
		*locations = append(*locations, LocationData{Text: location})
	}

	return IPData{
		StartIP:     startIP,
		EndIP:       endIP,
		LocationIdx: locIdx,
		prefix:      startIP >> 24,
	}, nil
}

// 从文件读取数据（支持 TXT 和 CSV）
func loadIPDataFromFile(filename string) ([]IPData, []LocationData, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, nil, fmt.Errorf("读取文件失败: %v", err)
	}

	lines := strings.Split(string(data), "\n")
	var ipDataList []IPData
	locationMap := make(map[string]uint32)
	var locations []LocationData

	isCSV := strings.HasSuffix(strings.ToLower(filename), ".csv")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var data IPData
		if isCSV {
			data, err = parseCSVData(line, locationMap, &locations)
		} else {
			data, err = parseIPData(line, locationMap, &locations)
		}
		if err != nil {
			fmt.Printf("解析错误: %v\n", err)
			continue
		}
		ipDataList = append(ipDataList, data)
	}
	return ipDataList, locations, nil
}

// 生成数据文件
func generateIPDat(filename string, ipDataList []IPData, locations []LocationData) error {
	sort.Slice(ipDataList, func(i, j int) bool {
		return ipDataList[i].StartIP < ipDataList[j].StartIP
	})

	prefixMap := make(map[uint32][]int)
	for i, ipData := range ipDataList {
		prefixMap[ipData.prefix] = append(prefixMap[ipData.prefix], i)
	}

	var buffer bytes.Buffer
	header := make([]byte, 16)
	prefixStartOffset := uint32(16)
	buffer.Write(header)

	indexCount := 0
	for prefix := uint32(0); prefix < 256; prefix++ {
		indices, exists := prefixMap[prefix]
		var startIndex, endIndex uint32
		if exists && len(indices) > 0 {
			startIndex = uint32(indexCount)
			endIndex = startIndex + uint32(len(indices)) - 1
			indexCount += len(indices)
		} else {
			startIndex = 0
			endIndex = 0
		}
		prefixBytes := []byte{byte(prefix)}
		startBytes := make([]byte, 4)
		endBytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(startBytes, startIndex)
		binary.LittleEndian.PutUint32(endBytes, endIndex)
		buffer.Write(prefixBytes)
		buffer.Write(startBytes)
		buffer.Write(endBytes)
	}
	prefixEndOffset := uint32(buffer.Len()) - 1
	firstStartIpOffset := prefixEndOffset + 1

	dataOffset := firstStartIpOffset + uint32(len(ipDataList)*12)
	for i := range locations {
		locations[i].Offset = dataOffset
		locations[i].Length = uint32(len(locations[i].Text))
		dataOffset += locations[i].Length
	}
	for _, ipData := range ipDataList {
		startIPBytes := make([]byte, 4)
		endIPBytes := make([]byte, 4)
		localOffsetBytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(startIPBytes, ipData.StartIP)
		binary.LittleEndian.PutUint32(endIPBytes, ipData.EndIP)
		binary.LittleEndian.PutUint32(localOffsetBytes, locations[ipData.LocationIdx].Offset)
		localLength := byte(locations[ipData.LocationIdx].Length)
		if localLength == 0 {
			localLength = 1
		}
		buffer.Write(startIPBytes)
		buffer.Write(endIPBytes)
		buffer.Write(localOffsetBytes[:3])
		buffer.WriteByte(localLength)
	}

	for _, loc := range locations {
		if loc.Text == "" {
			buffer.WriteString("|")
		} else {
			buffer.WriteString(loc.Text)
		}
	}

	result := buffer.Bytes()
	binary.LittleEndian.PutUint32(result[0:4], firstStartIpOffset)
	binary.LittleEndian.PutUint32(result[8:12], prefixStartOffset)
	binary.LittleEndian.PutUint32(result[12:16], prefixEndOffset)
	return os.WriteFile(filename, result, 0644)
}

// ipToUint32 将 IP 字符串转换为 uint32
func ipToUint32(ip string) uint32 {
	quads := strings.Split(ip, ".")
	var result uint32
	for i, q := range quads {
		n, _ := strconv.Atoi(q)
		result |= uint32(n) << (24 - i*8)
	}
	return result
}
