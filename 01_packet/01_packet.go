package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

type EthernetStr struct {
	DstEth  string
	SrcEth  string
	TypeEth uint16
}
type IPv4Packet struct {
	Version       uint8
	IHLBytes      uint8
	TotalLen      uint16
	ID            uint16
	Flags         string
	FragOffset    uint16
	TTL           uint8
	Protocol      uint8
	ChecksumValid bool
	Src           string
	Dst           string
}

type TCPSegment struct {
	SrcPort         uint16
	DstPort         uint16
	Seq             uint32
	Ack             uint32
	DataOffsetBytes uint8
	Flags           string
	Window          uint16
}

type UDPDatagram struct {
	SrcPort uint16
	DstPort uint16
	Length  uint16
}

func parseEthernet(data []byte) (*EthernetStr, error) {
	if len(data) < 14 {
		return nil, fmt.Errorf("кадр короче 20 байт, получено %d", len(data))
	}

	dstEth := fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",
		data[0], data[1], data[2], data[3], data[4], data[5])
	srcEth := fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",
		data[6], data[7], data[8], data[9], data[10], data[11])

	typeEth := binary.BigEndian.Uint16(data[12:14])
	return &EthernetStr{
		DstEth:  dstEth,
		SrcEth:  srcEth,
		TypeEth: typeEth,
	}, nil
}

func IPv4Parse(data []byte) (*IPv4Packet, error) {
	if len(data) < 20 {
		return nil, fmt.Errorf("кадр короче 20 байт, получено %d", len(data))
	}
	versionAndIHL := data[0]
	version := versionAndIHL >> 4
	if version != 4 {
		return nil, fmt.Errorf("версия IP равна %d, ожидалась 4", version)
	}
	ihlBytes := (versionAndIHL & 0x0F) * 4
	if ihlBytes < 20 {
		return nil, fmt.Errorf("длина заголовка IPv4 %d байт, минимум 20", ihlBytes)
	}
	if len(data) < int(ihlBytes) {
		return nil, fmt.Errorf("заголовок IPv4 занимает %d байт, а данных только %d", ihlBytes, len(data))
	}
	totalLen := binary.BigEndian.Uint16(data[2:4])
	id := binary.BigEndian.Uint16(data[4:6])
	flagsAndOffset := binary.BigEndian.Uint16(data[6:8])
	fragOffset := (flagsAndOffset & 0x1FFF) << 3
	flags := flagsAndOffset >> 13
	df := (flags & 2) != 0
	mf := (flags & 1) != 0
	var flagStr string
	if df && mf {
		flagStr = "DF,MF"
	} else if df {
		flagStr = "DF"
	} else if mf {
		flagStr = "MF"
	} else {
		flagStr = "none"
	}
	ttl := data[8]
	protocol := data[9]
	sum := uint32(0)
	for i := 0; i < int(ihlBytes); i += 2 {
		sum += uint32(binary.BigEndian.Uint16(data[i : i+2]))
	}
	for sum>>16 > 0 {
		sum = (sum & 0xFFFF) + (sum >> 16)
	}
	checksumValid := sum == 0xFFFF

	src := fmt.Sprintf("%d.%d.%d.%d", data[12], data[13], data[14], data[15])
	dst := fmt.Sprintf("%d.%d.%d.%d", data[16], data[17], data[18], data[19])
	return &IPv4Packet{
		Version:       version,
		IHLBytes:      ihlBytes,
		TotalLen:      totalLen,
		ID:            id,
		Flags:         flagStr,
		FragOffset:    fragOffset,
		TTL:           ttl,
		Protocol:      protocol,
		ChecksumValid: checksumValid,
		Src:           src,
		Dst:           dst,
	}, nil
}
func parseTCP(data []byte) (*TCPSegment, error) {
	dataOffBytes := (data[12] >> 4) * 4
	if len(data) < 20 {
		return nil, fmt.Errorf("кадр короче 20 байт, получено %d", len(data))
	}
	if len(data) < int(dataOffBytes) {
		return nil, fmt.Errorf("заголовок TCP занимает %d байт, а данных только %d", dataOffBytes, len(data))
	}
	srcPort := binary.BigEndian.Uint16(data[0:2])
	dstPort := binary.BigEndian.Uint16(data[2:4])
	seq := binary.BigEndian.Uint32(data[4:8])
	ackNum := binary.BigEndian.Uint32(data[8:12])
	flagsByte := data[13]
	names := []string{"FIN", "SYN", "RST", "PSH", "ACK", "URG"}
	var flagsList []string
	for i, name := range names {
		if flagsByte>>i&1 != 0 {
			flagsList = append(flagsList, name)
		}
	}
	flagsStr := "none"
	if len(flagsList) > 0 {
		flagsStr = strings.Join(flagsList, ",")
	}
	windows := binary.BigEndian.Uint16(data[14:16])
	return &TCPSegment{
		SrcPort:         srcPort,
		DstPort:         dstPort,
		Seq:             seq,
		Ack:             ackNum,
		DataOffsetBytes: dataOffBytes,
		Flags:           flagsStr,
		Window:          windows,
	}, nil
}

func parseUDP(data []byte) (*UDPDatagram, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("кадр короче 8 байт, получено %d", len(data))
	}
	src := binary.BigEndian.Uint16(data[0:2])
	dst := binary.BigEndian.Uint16(data[2:4])
	length := binary.BigEndian.Uint16(data[4:6])
	return &UDPDatagram{
		SrcPort: src,
		DstPort: dst,
		Length:  length,
	}, nil
}

func main() {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	input := string(data)

	var sb strings.Builder
	for _, r := range input {
		if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F') {
			sb.WriteRune(r)
		}
	}
	bytes, err := hex.DecodeString(sb.String())
	if err != nil {
		log.Fatal(err)
	}
	frame, err := parseEthernet(bytes)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("eth.dst %s\n", frame.DstEth)
	fmt.Printf("eth.src %s\n", frame.SrcEth)
	fmt.Printf("eth.ethertype 0x%04x\n", frame.TypeEth)

	if frame.TypeEth != 0x0800 {
		return
	}
	ip, err := IPv4Parse(bytes[14:])
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("ip.version %d\n", ip.Version)
	fmt.Printf("ip.ihl_bytes %d\n", ip.IHLBytes)
	fmt.Printf("ip.total_length %d\n", ip.TotalLen)
	fmt.Printf("ip.id 0x%04x\n", ip.ID)
	fmt.Printf("ip.flags %s\n", ip.Flags)
	fmt.Printf("ip.frag_offset %d\n", ip.FragOffset)
	fmt.Printf("ip.ttl %d\n", ip.TTL)
	fmt.Printf("ip.protocol %d\n", ip.Protocol)
	fmt.Printf("ip.checksum_valid %v\n", ip.ChecksumValid)
	fmt.Printf("ip.src %s\n", ip.Src)
	fmt.Printf("ip.dst %s\n", ip.Dst)

	tcpUdpStart := 14 + int(ip.IHLBytes)
	payloadStart := tcpUdpStart
	if ip.Protocol == 6 {
		tcp, err := parseTCP(bytes[tcpUdpStart:])
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("tcp.src_port %d\n", tcp.SrcPort)
		fmt.Printf("tcp.dst_port %d\n", tcp.DstPort)
		fmt.Printf("tcp.seq %d\n", tcp.Seq)
		fmt.Printf("tcp.ack %d\n", tcp.Ack)
		fmt.Printf("tcp.data_offset_bytes %d\n", tcp.DataOffsetBytes)
		fmt.Printf("tcp.flags %s\n", tcp.Flags)
		fmt.Printf("tcp.window %d\n", tcp.Window)

		payloadStart = tcpUdpStart + int(tcp.DataOffsetBytes)
	} else if ip.Protocol == 17 {
		udp, err := parseUDP(bytes[tcpUdpStart:])
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("udp.src_port %d\n", udp.SrcPort)
		fmt.Printf("udp.dst_port %d\n", udp.DstPort)
		fmt.Printf("udp.length %d\n", udp.Length)

		payloadStart = tcpUdpStart + 8

	}
	payloadLength := (14 + int(ip.TotalLen)) - payloadStart
	if payloadLength < 0 {
		payloadLength = 0
	}
	fmt.Printf("payload.length %d\n", payloadLength)

}
