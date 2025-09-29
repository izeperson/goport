// GoPort network port scanner. (cool right?)
package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"sort"
	"sync"
	"time"
)

func portScan(target string, start, end, concurrency int, timeout time.Duration) []int { // checks machine for open ports
	ports := make(chan int)
	results := make(chan int)
	var wg sync.WaitGroup
	worker := func() {
		defer wg.Done()
		for p := range ports {
			addr := net.JoinHostPort(target, fmt.Sprint(p))
			conn, err := net.DialTimeout("tcp", addr, timeout)
			if err == nil {
				conn.Close()
				results <- p
			}
		}
	}
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go worker()
	}
	go func() {
		for p := start; p <= end; p++ {
			ports <- p
		}
		close(ports)
	}()
	open := []int{}
	go func() {
		wg.Wait()
		close(results)
	}()
	for p := range results {
		open = append(open, p)
	}
	sort.Ints(open)
	return open
}

func pingFunc(target string, port int, timeout time.Duration, count int, interval time.Duration) (int, int, int, int) {
	success := 0
	min := int(^uint(0) >> 1)
	max := 0
	total := 0
	for i := 0; i < count; i++ {
		start := time.Now()
		addr := net.JoinHostPort(target, fmt.Sprint(port))
		conn, err := net.DialTimeout("tcp", addr, timeout)
		d := int(time.Since(start).Milliseconds())
		if err == nil {
			success++
			conn.Close()
			if d < min {
				min = d
			}
			if d > max {
				max = d
			}
			total += d
		}
		if i < count-1 {
			time.Sleep(interval)
		}
	}
	if min == int(^uint(0)>>1) {
		min = 0
	}
	return success, min, max, total
}

func help() { // hurr hurr i need help because i cant read code. hurrrrrr
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  -mode scan|ping")
	fmt.Fprintln(os.Stderr, "Scan options:")
	fmt.Fprintln(os.Stderr, "  -target host (required for both)")
	fmt.Fprintln(os.Stderr, "  -start n  (scan start port, default 1)")
	fmt.Fprintln(os.Stderr, "  -end n    (scan end port, default 65535)")
	fmt.Fprintln(os.Stderr, "  -concurrency n (default 100)")
	fmt.Fprintln(os.Stderr, "  -timeout ms (timeout in milliseconds, default 500)")
	fmt.Fprintln(os.Stderr, "Ping options:")
	fmt.Fprintln(os.Stderr, "  -port n   (port to ping for mode=ping)")
	fmt.Fprintln(os.Stderr, "  -count n  (amount of pings, default 4)")
	fmt.Fprintln(os.Stderr, "  -interval ms (interval between pings, default 1000)")
}

func main() { // checks options
	mode := flag.String("mode", "scan", "scan or ping")
	target := flag.String("target", "", "target host or IP")
	start := flag.Int("start", 1, "start port")
	end := flag.Int("end", 65535, "end port")
	concurrency := flag.Int("concurrency", 100, "concurrency")
	timeoutMs := flag.Int("timeout", 500, "timeout in ms")
	port := flag.Int("port", 0, "port for ping")
	count := flag.Int("count", 4, "ping count")
	intervalMs := flag.Int("interval", 1000, "ping interval ms")
	flag.Parse()
	if *target == "" {
		help()
		os.Exit(2) // if no flags are used, silent quit program.
	}
	timeout := time.Duration(*timeoutMs) * time.Millisecond
	if *mode == "scan" {
		if *start < 1 {
			*start = 1
		}
		if *end > 65535 { // end port range because tcp/ip is cool
			*end = 65535
		}
		if *start > *end {
			fmt.Fprintln(os.Stderr, "start must be <= end")
			os.Exit(2)
		}
		open := portScan(*target, *start, *end, *concurrency, timeout)
		for _, p := range open {
			fmt.Printf("%d open\n", p)
		}
		fmt.Printf("scan complete: %d open ports\n", len(open))
		return
	}
	if *mode == "ping" {
		if *port <= 0 || *port > 65535 {
			fmt.Fprintln(os.Stderr, "invalid port for ping")
			os.Exit(2)
		}
		interval := time.Duration(*intervalMs) * time.Millisecond
		success, min, max, total := pingFunc(*target, *port, timeout, *count, interval)
		avg := 0
		if success > 0 {
			avg = total / success
		}
		fmt.Printf("pings: %d/%d success\n", success, *count)
		fmt.Printf("min=%dms avg=%dms max=%dms\n", min, avg, max)
		return
	}
	help()
	os.Exit(2)
}
