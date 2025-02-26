package ip2asn

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

// ipData 表示一条 IP 范围和对应的 ASN 信息
type ipData struct {
	StartIP uint32 // 起始 IP
	EndIP   uint32 // 结束 IP
	ASNIdx  uint32 // ASN 信息在数据区中的索引
	prefix  uint32 // 前缀（IP 的第一个八位字节）
}

// asnData 表示去重后的 ASN 信息
type asnData struct {
	Offset uint32 // 在数据区中的偏移量
	Length uint32 // 信息的长度
	Text   string // ASN 信息字符串（ipRange|asn|组织名称）
}

func Convert(inputFile, outputFile string) (err error) {
	ipDataList, asnList, err := loadIPDataFromFile(inputFile)
	if err != nil {
		fmt.Println("加载数据失败:", err)
		return err
	}
	err = generateIPDat(outputFile, ipDataList, asnList)
	if err != nil {
		fmt.Println("生成文件失败:", err)
		return err
	}
	fmt.Printf("生成文件成功: %s\n", outputFile)
	return
}

// 从文本行解析 CSV 格式的 ipData 和 asnData
func parseCSVData(line string, asnMap map[string]uint32, asnList *[]asnData) (ipData, error) {
	fields := strings.Split(line, ",")
	if len(fields) < 5 { // 需要 5 个字段：startIPNum, endIPNum, ipRange, asn, org
		return ipData{}, fmt.Errorf("CSV 字段不足: %s", line)
	}

	// 移除引号并填充到 5 个字段
	for i, field := range fields {
		fields[i] = strings.Trim(field, `"`)
	}
	for len(fields) < 5 {
		fields = append(fields, "")
	}

	// 解析 startIPNum 和 endIPNum 为 uint32
	startIP, err := strconv.ParseUint(fields[0], 10, 32)
	if err != nil {
		return ipData{}, fmt.Errorf("无效的起始 IP: %s", fields[0])
	}
	endIP, err := strconv.ParseUint(fields[1], 10, 32)
	if err != nil {
		return ipData{}, fmt.Errorf("无效的结束 IP: %s", fields[1])
	}

	// 拼接 ASN 信息（ipRange|asn|org）
	asnInfo := strings.Join(fields[2:5], "|")

	var asnIdx uint32
	if offset, exists := asnMap[asnInfo]; exists {
		asnIdx = offset
	} else {
		asnIdx = uint32(len(*asnList))
		asnMap[asnInfo] = asnIdx
		*asnList = append(*asnList, asnData{Text: asnInfo})
	}

	return ipData{
		StartIP: uint32(startIP),
		EndIP:   uint32(endIP),
		ASNIdx:  asnIdx,
		prefix:  uint32(startIP >> 24),
	}, nil
}

// 从文件读取数据（仅支持 CSV）
func loadIPDataFromFile(filename string) ([]ipData, []asnData, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, nil, fmt.Errorf("读取文件失败: %v", err)
	}

	lines := strings.Split(string(data), "\n")
	var ipDataList []ipData
	asnMap := make(map[string]uint32)
	var asns []asnData

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		data, err := parseCSVData(line, asnMap, &asns)
		if err != nil {
			fmt.Printf("解析错误: %v\n", err)
			continue
		}
		ipDataList = append(ipDataList, data)
	}
	return ipDataList, asns, nil
}

// 生成数据文件
func generateIPDat(filename string, ipDataList []ipData, asns []asnData) error {
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
	for i := range asns {
		asns[i].Offset = dataOffset
		asns[i].Length = uint32(len(asns[i].Text))
		dataOffset += asns[i].Length
		if asns[i].Length > 255 {
			fmt.Printf("警告：ASN信息长度超255字节：%d\n", asns[i].Length)
		}
	}

	for _, ipData := range ipDataList {
		startIPBytes := make([]byte, 4)
		endIPBytes := make([]byte, 4)
		localOffsetBytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(startIPBytes, ipData.StartIP)
		binary.LittleEndian.PutUint32(endIPBytes, ipData.EndIP)
		binary.LittleEndian.PutUint32(localOffsetBytes, asns[ipData.ASNIdx].Offset)
		localLength := byte(asns[ipData.ASNIdx].Length)
		if localLength == 0 {
			localLength = 1
		}
		buffer.Write(startIPBytes)
		buffer.Write(endIPBytes)
		buffer.Write(localOffsetBytes) // 4字节偏移
		buffer.WriteByte(localLength)
	}

	// 内容区
	for _, asn := range asns {
		if asn.Text == "" {
			buffer.WriteString("|")
		} else {
			buffer.WriteString(asn.Text)
		}
	}

	result := buffer.Bytes()
	binary.LittleEndian.PutUint32(result[0:4], firstStartIpOffset)
	binary.LittleEndian.PutUint32(result[8:12], prefixStartOffset)
	binary.LittleEndian.PutUint32(result[12:16], prefixEndOffset)
	fmt.Printf("生成文件大小: %d 字节\n", len(result))
	return os.WriteFile(filename, result, 0644)
}
