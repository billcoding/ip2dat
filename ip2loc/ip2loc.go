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

type ipData struct {
	StartIP     uint32
	EndIP       uint32
	LocationIdx uint32
	prefix      uint32
}

type locationData struct {
	Offset uint32
	Length uint32
	Text   string
}

func Convert(inputFile, outputFile string) (err error) {
	ipDataList, locations, err := loadIPDataFromFile(inputFile)
	if err != nil {
		fmt.Println("加载数据失败:", err)
		return err
	}
	err = generateIPDat(outputFile, ipDataList, locations)
	if err != nil {
		fmt.Println("生成文件失败:", err)
		return err
	}
	fmt.Printf("生成文件成功: %s\n", outputFile)
	return
}

func loadIPDataFromFile(filename string) ([]ipData, []locationData, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, nil, fmt.Errorf("读取文件失败: %v", err)
	}

	lines := strings.Split(string(data), "\n")
	var ipDataList []ipData
	locationMap := make(map[string]uint32)
	var locations []locationData

	isCSV := strings.HasSuffix(strings.ToLower(filename), ".csv")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// 可选：限制行数测试
		// if i >= 1000 { break } // 测试小文件时启用
		var data ipData
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

func parseIPData(line string, locationMap map[string]uint32, locations *[]locationData) (ipData, error) {
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
		*locations = append(*locations, locationData{Text: location})
	}

	return ipData{StartIP: startIP, EndIP: endIP, LocationIdx: locIdx, prefix: startIP >> 24}, nil
}

func parseCSVData(line string, locationMap map[string]uint32, locations *[]locationData) (ipData, error) {
	fields := strings.Split(line, ",")
	if len(fields) < 15 {
		for len(fields) < 15 {
			fields = append(fields, "")
		}
	}

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
		*locations = append(*locations, locationData{Text: location})
	}

	return ipData{StartIP: startIP, EndIP: endIP, LocationIdx: locIdx, prefix: startIP >> 24}, nil
}

func generateIPDat(filename string, ipDataList []ipData, locations []locationData) error {
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

	// 前缀区：256 * 9字节
	for prefix := uint32(0); prefix < 256; prefix++ {
		indices, exists := prefixMap[prefix]
		var startIndex, endIndex uint32
		if exists && len(indices) > 0 {
			startIndex = uint32(indices[0])
			endIndex = uint32(indices[len(indices)-1])
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

	// 索引区：13字节每条（4字节偏移）
	dataOffset := firstStartIpOffset + uint32(len(ipDataList)*13)
	for i := range locations {
		locations[i].Offset = dataOffset
		locations[i].Length = uint32(len(locations[i].Text))
		dataOffset += locations[i].Length
		if locations[i].Length > 255 {
			fmt.Printf("警告：地理信息长度超255字节：%d\n", locations[i].Length)
		}
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
		buffer.Write(localOffsetBytes) // 4字节偏移
		buffer.WriteByte(localLength)
	}

	// 内容区
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
	fmt.Printf("生成文件大小: %d 字节\n", len(result))
	return os.WriteFile(filename, result, 0644)
}

func ipToUint32(ip string) uint32 {
	quads := strings.Split(ip, ".")
	var result uint32
	for i, q := range quads {
		n, _ := strconv.Atoi(q)
		result |= uint32(n) << (24 - i*8)
	}
	return result
}
