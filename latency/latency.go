package main

import (
	"fmt"
	"os"
	"time"
	ping "github.com/sparrc/go-ping"
)

type Message struct {
	rtt time.Duration
//	pkt_sent int
//	pkt_rcvd int
}

const (
	N = 60
)

var	(
	err error
	alpha float32 = 0
	latency_x_1 time.Duration
/*	avg_latency_x_1 float32 = 0
	avg_variance_latency_x_1 float32 = 0
	jitter_x_1 float32 = 0
	avg_jitter_x_1 float32 = 0
	avg_variance_jetter_x_1 float32 = 0*/
)

func ping_it(host string) (*Message, error) {
	pinger, err := ping.NewPinger(host)
	if err != nil {
		return nil, err
	}

	pinger.Count = 1
	pinger.Run() // blocks until finished
	stats := pinger.Statistics() // get send/receive/rtt stats
	fmt.Println("PacketsRecv = ", stats.PacketsRecv)
	fmt.Println("PacketsSent = ", stats.PacketsSent)
	fmt.Println("PacketLoss  = ", stats.PacketLoss)
	fmt.Println("IPAddr      = ", stats.IPAddr)
	fmt.Println("Addr        = ", stats.Addr)
	fmt.Println("Rtts        = ", stats.Rtts)
	fmt.Println("MinRtt      = ", stats.MinRtt)
	fmt.Println("MaxRtt      = ", stats.MaxRtt)
	fmt.Println("AvgRtt      = ", stats.AvgRtt)
	fmt.Println("StdDevRtt   = ", stats.StdDevRtt)
	return &Message{
        rtt: stats.AvgRtt,
    }, nil
}
/*
func latency(host string) {
	// Initialize values
	alpha = 2 / (N+1)
	latency_x_1, err = ping_it(host)
	fmt.Println("alpha = ", alpha)
}
*/
// ping
func main() {
	m Message{}
	m, err = ping_it("www.google.com")
	fmt.Println(err)
	fmt.Println(os.Stdout)
	fmt.Println("latency = ", latency_x_1)
}
