package ipasnsearch

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type ipIndex struct {
	startIp, endIp, localOffset, localLength uint32
}
type prefixIndex struct {
	startIndex, endIndex uint32
}

type Searcher struct {
	data      []byte
	prefixMap map[uint32]prefixIndex
	firstStartIpOffset,
	prefixStartOffset,
	prefixEndOffset,
	prefixCount uint32
}

func Search(datFile, ip string) string {
	s, err := New(datFile)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return s.Get(ip)
}

func New(datFile string) (*Searcher, error) {
	s := Searcher{}
	data, err := os.ReadFile(datFile)
	if err != nil {
		log.Fatal(err)
	}
	s.data = data
	s.prefixMap = make(map[uint32]prefixIndex)

	s.firstStartIpOffset = bytesToLong(data[0], data[1], data[2], data[3])
	s.prefixStartOffset = bytesToLong(data[8], data[9], data[10], data[11])
	s.prefixEndOffset = bytesToLong(data[12], data[13], data[14], data[15])
	s.prefixCount = (s.prefixEndOffset-s.prefixStartOffset)/9 + 1

	indexBuffer := s.data[s.prefixStartOffset:(s.prefixEndOffset + 9)]
	for k := uint32(0); k < s.prefixCount; k++ {
		i := k * 9
		prefix := uint32(indexBuffer[i] & 0xFF)
		pf := prefixIndex{}
		pf.startIndex = bytesToLong(indexBuffer[i+1], indexBuffer[i+2], indexBuffer[i+3], indexBuffer[i+4])
		pf.endIndex = bytesToLong(indexBuffer[i+5], indexBuffer[i+6], indexBuffer[i+7], indexBuffer[i+8])
		s.prefixMap[prefix] = pf
	}
	return &s, nil
}

func (s *Searcher) Get(ip string) string {
	ipS := strings.Split(ip, ".")
	x, _ := strconv.Atoi(ipS[0])
	prefix := uint32(x)
	intIP := ipToLong(ip)

	var high, low uint32
	if pf, ok := s.prefixMap[prefix]; ok {
		low = pf.startIndex
		high = pf.endIndex
	} else {
		return ""
	}

	var myIndex uint32
	if low == high {
		myIndex = low
	} else {
		myIndex = s.binarySearch(low, high, intIP)
	}

	index := ipIndex{}
	index.getIndex(myIndex, s)

	if index.startIp <= intIP && index.endIp >= intIP {
		return index.getLocal(s)
	}
	return ""
}

func (s *Searcher) binarySearch(low, high, k uint32) uint32 {
	var M uint32
	for low <= high {
		mid := (low + high) / 2
		endIpNum := s.getEndIp(mid)
		if endIpNum >= k {
			M = mid
			if mid == 0 {
				break
			}
			high = mid - 1
		} else {
			low = mid + 1
		}
	}
	return M
}

func (s *Searcher) getEndIp(left uint32) uint32 {
	leftOffset := s.firstStartIpOffset + left*13
	return bytesToLong(s.data[4+leftOffset], s.data[5+leftOffset], s.data[6+leftOffset], s.data[7+leftOffset])
}

func (p *ipIndex) getIndex(left uint32, ips *Searcher) {
	leftOffset := ips.firstStartIpOffset + left*13
	p.startIp = bytesToLong(ips.data[leftOffset], ips.data[1+leftOffset], ips.data[2+leftOffset], ips.data[3+leftOffset])
	p.endIp = bytesToLong(ips.data[4+leftOffset], ips.data[5+leftOffset], ips.data[6+leftOffset], ips.data[7+leftOffset])
	p.localOffset = bytesToLong(ips.data[8+leftOffset], ips.data[9+leftOffset], ips.data[10+leftOffset], ips.data[11+leftOffset])
	p.localLength = uint32(ips.data[12+leftOffset])
}

func (p *ipIndex) getLocal(ips *Searcher) string {
	bytes := ips.data[p.localOffset : p.localOffset+p.localLength]
	return string(bytes)
}

func ipToLong(ip string) uint32 {
	quads := strings.Split(ip, ".")
	var result uint32
	for i, q := range quads {
		n, _ := strconv.Atoi(q)
		result |= uint32(n) << (24 - i*8)
	}
	return result
}

func bytesToLong(a, b, c, d byte) uint32 {
	return uint32(a) | (uint32(b) << 8) | (uint32(c) << 16) | (uint32(d) << 24)
}

func main() {
	// 示例使用
	result := Search("asn.dat", "8.8.8.8")
	fmt.Println(result)
}
