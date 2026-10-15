package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func fail(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}

func parseIP(s string) (uint32, error) {
	parts := strings.Split(s, ".")
	if len(parts) != 4 {
		return 0, fmt.Errorf("bad ip: %s", s)
	}
	var res uint32
	for _, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 || n > 255 {
			return 0, fmt.Errorf("bad ip: %s", s)
		}
		res = res<<8 | uint32(n)
	}
	return res, nil
}

func ipToString(x uint32) string {
	return fmt.Sprintf("%d.%d.%d.%d", byte(x>>24), byte(x>>16), byte(x>>8), byte(x))
}

func parseCIDR(s string) (uint32, int, error) {
	parts := strings.Split(s, "/")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("bad cidr: %s", s)
	}
	ip, err := parseIP(parts[0])
	if err != nil {
		return 0, 0, err
	}
	n, err := strconv.Atoi(parts[1])
	if err != nil || n < 0 || n > 32 {
		return 0, 0, fmt.Errorf("bad prefix: %s", s)
	}
	return ip, n, nil
}

func maskFor(n int) uint32 {
	if n == 0 {
		return 0
	}
	return ^uint32(0) << (32 - n)
}

func runSubnet(args []string) {
	if len(args) != 1 {
		fail("usage: subnet <addr>/<prefix>")
	}
	ip, n, err := parseCIDR(args[0])
	if err != nil {
		fail(err.Error())
	}
	mask := maskFor(n)
	network := ip & mask
	bcast := network | ^mask

	var first, last uint32
	var broadcast string
	var hosts uint64

	switch n {
	case 32:
		first, last = network, network
		broadcast = "none"
		hosts = 1
	case 31:
		first, last = network, network+1
		broadcast = "none"
		hosts = 2
	default:
		first, last = network+1, bcast-1
		broadcast = ipToString(bcast)
		hosts = (uint64(1) << uint(32-n)) - 2
	}

	fmt.Printf("network %s\n", ipToString(network))
	fmt.Printf("broadcast %s\n", broadcast)
	fmt.Printf("netmask %s\n", ipToString(mask))
	fmt.Printf("prefix %d\n", n)
	fmt.Printf("first %s\n", ipToString(first))
	fmt.Printf("last %s\n", ipToString(last))
	fmt.Printf("hosts %d\n", hosts)
}

func runRoute(args []string) {
	if len(args) != 2 {
		fail("usage: route <file> <addr>")
	}
	dest, err := parseIP(args[1])
	if err != nil {
		fail(err.Error())
	}
	f, err := os.Open(args[0])
	if err != nil {
		fail(err.Error())
	}
	defer f.Close()

	best := -1
	bestIface := ""

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		netw, n, err := parseCIDR(fields[0])
		if err != nil {
			continue
		}
		mask := maskFor(n)
		if dest&mask == netw&mask && n > best {
			best = n
			bestIface = fields[1]
		}
	}

	if best == -1 {
		fmt.Println("unreachable true")
		os.Exit(1)
	}
	fmt.Printf("via %s\n", bestIface)
	fmt.Printf("prefix %d\n", best)
}

func main() {
	if len(os.Args) < 2 {
		fail("usage: subnet|route ...")
	}
	switch os.Args[1] {
	case "subnet":
		runSubnet(os.Args[2:])
	case "route":
		runRoute(os.Args[2:])
	default:
		fail("unknown mode: " + os.Args[1])
	}
}
