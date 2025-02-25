package ipasnsearch

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type (
	ipIndex struct {
		startIp, endIp, localOffset, localLength uint32
	}
	prefixIndex struct {
		startIndex, endIndex uint32
	}
)

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

	var high uint32 = 0
	var low uint32 = 0

	if _, ok := s.prefixMap[prefix]; ok {
		low = s.prefixMap[prefix].startIndex
		high = s.prefixMap[prefix].endIndex
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
		result := index.getLocal(s)
		return result
	}
	return ""
}

func (s *Searcher) binarySearch(low uint32, high uint32, k uint32) uint32 {
	var M uint32 = 0
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
	leftOffset := s.firstStartIpOffset + left*12
	return bytesToLong(s.data[4+leftOffset], s.data[5+leftOffset], s.data[6+leftOffset], s.data[7+leftOffset])
}

func (p *ipIndex) getIndex(left uint32, s *Searcher) {
	leftOffset := s.firstStartIpOffset + left*12
	p.startIp = bytesToLong(s.data[leftOffset], s.data[1+leftOffset], s.data[2+leftOffset], s.data[3+leftOffset])
	p.endIp = bytesToLong(s.data[4+leftOffset], s.data[5+leftOffset], s.data[6+leftOffset], s.data[7+leftOffset])
	p.localOffset = bytesToLong3(s.data[8+leftOffset], s.data[9+leftOffset], s.data[10+leftOffset])
	p.localLength = uint32(s.data[11+leftOffset])
}

func (p *ipIndex) getLocal(s *Searcher) string {
	bytes := s.data[p.localOffset : p.localOffset+p.localLength]
	return string(bytes)
}

func ipToLong(ip string) uint32 {
	quads := strings.Split(ip, ".")
	var result uint32 = 0
	a, _ := strconv.Atoi(quads[3])
	result += uint32(a)
	b, _ := strconv.Atoi(quads[2])
	result += uint32(b) << 8
	c, _ := strconv.Atoi(quads[1])
	result += uint32(c) << 16
	d, _ := strconv.Atoi(quads[0])
	result += uint32(d) << 24
	return result
}

func bytesToLong(a, b, c, d byte) uint32 {
	a1 := uint32(a)
	b1 := uint32(b)
	c1 := uint32(c)
	d1 := uint32(d)
	return (a1 & 0xFF) | ((b1 << 8) & 0xFF00) | ((c1 << 16) & 0xFF0000) | ((d1 << 24) & 0xFF000000)
}

func bytesToLong3(a, b, c byte) uint32 {
	a1 := uint32(a)
	b1 := uint32(b)
	c1 := uint32(c)
	return (a1 & 0xFF) | ((b1 << 8) & 0xFF00) | ((c1 << 16) & 0xFF0000)
}
