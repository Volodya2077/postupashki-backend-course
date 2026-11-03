package main

import (
	"bufio"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

const (
	typeA     uint16 = 1
	typeNS    uint16 = 2
	typeCNAME uint16 = 5
	typeMX    uint16 = 15
	typeTXT   uint16 = 16
	typeAAAA  uint16 = 28
)

type Record struct {
	Type  string
	Value string
	TTL   uint32
}

type cacheEntry struct {
	record  []Record
	expires time.Time
}

func buildQuery(name string, qtype uint16) ([]byte, uint16, error) {
	result := []byte{}
	header := []byte{
		0x00, 0x00,
		0x01, 0x00,
		0x00, 0x01,
		0x00, 0x00,
		0x00, 0x00,
		0x00, 0x00,
	}
	var idBytes [2]byte
	if _, err := rand.Read(idBytes[:]); err != nil {
		return nil, 0, err
	}
	id := binary.BigEndian.Uint16(idBytes[:])

	binary.BigEndian.PutUint16(header[0:2], id)

	en, err := encodeName(name)
	if err != nil {
		return nil, 0, err
	}
	typeBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(typeBytes, qtype)

	result = append(result, header...)
	result = append(result, en...)
	result = append(result, typeBytes...)
	result = append(result, 0x00, 0x01)

	return result, id, nil
}

func encodeName(name string) ([]byte, error) {
	if name == "." {
		return []byte{0x00}, nil
	}
	name = strings.TrimSuffix(name, ".")
	parts := strings.Split(name, ".")
	result := []byte{}
	for _, part := range parts {
		if len(part) < 1 || len(part) > 63 {
			return nil, fmt.Errorf("длина метки %d вне диапазона 1..63", len(part))
		}

		result = append(result, byte(len(part)))
		result = append(result, []byte(part)...)
	}
	result = append(result, 0x00)
	if len(result) > 255 {
		return nil, errors.New("имя длиннее 255 байт")
	}
	return result, nil

}

func readResponse(conn net.Conn, id uint16) ([]byte, error) {
	buf := make([]byte, 4096)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	for {
		n, err := conn.Read(buf)
		if err != nil {
			return nil, err
		}
		if n < 12 {
			continue
		}
		if binary.BigEndian.Uint16(buf[0:2]) != id {
			continue
		}
		return buf[:n:n], nil
	}
}

func rcodeName(rcode uint16) string {
	switch rcode {
	case 0:
		return "NOERROR"
	case 1:
		return "FORMERR"
	case 2:
		return "SERVFAIL"
	case 3:
		return "NXDOMAIN"
	case 5:
		return "REFUSED"
	default:
		return fmt.Sprintf("RCODE%d", rcode)
	}
}

func formatIPv6(data []byte) string {
	if len(data) != 16 {
		return "invalid"
	}

	groups := make([]uint16, 8)
	for i := 0; i < 8; i++ {
		groups[i] = binary.BigEndian.Uint16(data[i*2 : i*2+2])
	}

	maxStart, maxLen := -1, 0
	curStart, curLen := -1, 0
	for i := 0; i < 8; i++ {
		if groups[i] == 0 {
			if curStart == -1 {
				curStart, curLen = i, 1
			} else {
				curLen++
			}
			if curLen > maxLen {
				maxStart, maxLen = curStart, curLen
			}
		} else {
			curStart, curLen = -1, 0
		}
	}
	if maxLen < 2 {
		maxStart = -1
	}
	parts := []string{}
	i := 0
	for i < 8 {
		if i == maxStart {
			if i == 0 {
				parts = append(parts, "")
			}
			parts = append(parts, "")
			i += maxLen
			if i == 8 {
				parts = append(parts, "")
			}
		} else {
			parts = append(parts, fmt.Sprintf("%x", groups[i]))
			i++
		}
	}

	result := strings.Join(parts, ":")
	return result
}

func readName(msg []byte, offset int) (string, int, error) {
	labels := []string{}
	end := -1
	jumps := 0

	for {
		if offset < 0 || offset >= len(msg) {
			return "", 0, errors.New("смещение вне диапазона")
		}
		b := msg[offset]

		switch b & 0xC0 {
		case 0xC0:
			if offset+2 > len(msg) {
				return "", 0, errors.New("указатель усеченный")
			}
			jumps++
			if jumps > 128 {
				return "", 0, errors.New("зациклился указатель")
			}
			if end == -1 {
				end = offset + 2
			}
			offset = int(binary.BigEndian.Uint16(msg[offset:offset+2]) & 0x3FFF)
		case 0x00:
			if b == 0 {
				if end == -1 {
					end = offset + 1
				}
				return strings.Join(labels, ".") + ".", end, nil
			}
			l := int(b)
			if offset+1+l > len(msg) {
				return "", 0, errors.New("метка усеченная")

			}
			labels = append(labels, string(msg[offset+1:offset+1+l]))
			offset += 1 + l
		default:
			return "", 0, fmt.Errorf("зарезервированный тип метки 0x%02x", b&0xC0)

		}
	}

}

func parseAnswers(msg []byte) ([]Record, error) {
	qdcount := int(binary.BigEndian.Uint16(msg[4:6]))
	ancount := int(binary.BigEndian.Uint16(msg[6:8]))
	offset := 12
	var err error
	for i := 0; i < qdcount; i++ {
		_, offset, err = readName(msg, offset)
		if err != nil {
			return nil, err
		}
		offset += 4
		if offset > len(msg) {
			return nil, errors.New("вопрос усеченная")
		}
	}
	records := []Record{}
	for i := 0; i < ancount; i++ {
		_, offset, err = readName(msg, offset)
		if err != nil {
			return nil, err
		}
		if offset+10 > len(msg) {
			return nil, errors.New("заголовок записи усеченный")
		}
		rtype := binary.BigEndian.Uint16(msg[offset : offset+2])
		ttl := binary.BigEndian.Uint32(msg[offset+4 : offset+8])
		if ttl&0x80000000 != 0 {
			ttl = 0
		}
		rdlen := int(binary.BigEndian.Uint16(msg[offset+8 : offset+10]))
		offset += 10
		rdataOffset := offset
		rdataEnd := rdataOffset + rdlen

		if rdataEnd > len(msg) {
			return nil, errors.New("данные записи усеченные")
		}

		switch rtype {
		case typeA:
			if rdlen == 4 {
				value := net.IP(msg[rdataOffset : rdataOffset+4]).String()
				records = append(records, Record{"A", value, ttl})
			}

		case typeAAAA:
			if rdlen == 16 {
				value := formatIPv6(msg[rdataOffset : rdataOffset+16])
				records = append(records, Record{"AAAA", value, ttl})
			}
		case typeCNAME, typeNS:
			name, _, err := readName(msg, rdataOffset)
			if err != nil {
				return nil, err
			}
			typeName := "CNAME"
			if rtype == typeNS {
				typeName = "NS"

			}
			records = append(records, Record{typeName, name, ttl})
		case typeMX:
			if rdlen >= 2 {
				priority := binary.BigEndian.Uint16(msg[rdataOffset : rdataOffset+2])
				name, _, err := readName(msg, rdataOffset+2)
				if err != nil {
					return nil, err
				}
				value := fmt.Sprintf("%d %s", priority, name)
				records = append(records, Record{"MX", value, ttl})
			}
		case typeTXT:
			var sb strings.Builder
			pos := rdataOffset
			for pos < rdataOffset+rdlen {
				l := int(msg[pos])
				pos++
				if pos+l > rdataEnd {
					return nil, errors.New("строка TXT выходит за пределы записи")
				}

				sb.Write(msg[pos : pos+l])

				pos += l
			}
			records = append(records, Record{"TXT", sb.String(), ttl})

		}
		offset = rdataEnd
	}
	return records, nil
}

// добавил функцию тк убрал мапу
func typeCode(name string) (uint16, bool) {
	switch name {
	case "A":
		return typeA, true
	case "NS":
		return typeNS, true
	case "CNAME":
		return typeCNAME, true
	case "MX":
		return typeMX, true
	case "TXT":
		return typeTXT, true
	case "AAAA":
		return typeAAAA, true
	}
	return 0, false
}

func main() {
	addr := os.Args[1]
	port := os.Args[2]

	scanner := bufio.NewScanner(os.Stdin)
	conn, err := net.Dial("udp", net.JoinHostPort(addr, port))
	if err != nil {
		log.Fatal("ошибка в dial")
	}
	defer conn.Close()

	cache := map[string]cacheEntry{}

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		name := parts[0]
		typeName := parts[1]
		key := strings.ToLower(name) + " " + typeName
		code, ok := typeCode(typeName)
		if !ok {
			continue
		}
		query, id, err := buildQuery(name, code)
		if err != nil {
			continue
		}
		fmt.Println("query", name, typeName)

		entry, found := cache[key]

		if found && time.Now().Before(entry.expires) {
			fmt.Println("status NOERROR")
			for _, r := range entry.record {
				fmt.Println("answer", r.Type, r.Value, r.TTL)
			}
			fmt.Println("end")
			continue
		}

		conn.Write(query)

		response, err := readResponse(conn, id)
		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				fmt.Println("status TIMEOUT")
			} else {
				log.Println("ошибка чтения:", err)
				fmt.Println("status ERROR")
			}
			fmt.Println("end")
			os.Exit(1)
		}

		flags := binary.BigEndian.Uint16(response[2:4])
		rcode := flags & 0x0F
		records, err := parseAnswers(response)
		if err != nil {
			log.Println("malformed response:", err)
			fmt.Println("status FORMERR")
			fmt.Println("end")
			continue
		}

		fmt.Println("status", rcodeName(rcode))
		for _, r := range records {
			fmt.Println("answer", r.Type, r.Value, r.TTL)
		}
		if rcode == 0 && len(records) > 0 {
			minTTL := records[0].TTL
			for _, r := range records {
				if r.TTL < minTTL {
					minTTL = r.TTL
				}
			}
			if minTTL > 0 {
				cache[key] = cacheEntry{records, time.Now().Add(time.Duration(minTTL) * time.Second)}
			}
		}
		fmt.Println("end")

	}

}
