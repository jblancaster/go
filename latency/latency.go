package main

import (
	"fmt"
//	"math"
	"os"
	"time"
	ping "github.com/sparrc/go-ping"
)

type Message struct {
	rtt time.Duration
	pkt_sent int
	pkt_rcvd int
}

const (
	N = 60
)

var	(
	err error
	alpha float32 = 0
	jitter_x   time.Duration
//	jitter_x   float32 = 0
//	jitter_x_1 float32 = 0
	latency_x   time.Duration
	latency_x_1 time.Duration
//	avg_latency_x   time.Duration
	avg_latency_x_1 time.Duration
/*	avg_jitter_x    float32 = 0
	avg_jitter_x_1  float32 = 0
	avg_variance_latency_x   float32 = 0
	avg_variance_latency_x_1 float32 = 0
	avg_variance_jitter_x    float32 = 0
	avg_variance_jitter_x_1  float32 = 0
*/
)

func ping_it(host string) (*Message, error) {
	pinger, err := ping.NewPinger(host)
	if err != nil {
		return nil, err
	}

	pinger.Count = 1
	pinger.Run() // blocks until finished
	stats := pinger.Statistics() // get send/receive/rtt stats
	return &Message {
        rtt:      stats.AvgRtt,
        pkt_sent: stats.PacketsSent,
        pkt_rcvd: stats.PacketsRecv,
    }, nil
}

// ping
func main() {

	m, err := ping_it("www.google.com")

	// Initialize variables
	latency_x_1 = m.rtt
	avg_latency_x_1 = latency_x_1

	// Get new data point
	m, err = ping_it("www.google.com")
	latency_x = m.rtt
	jitter_x = latency_x - latency_x_1
//	latencyInMiliSeconds := int64(jitter / time.Microsecond)

	fmt.Println(err)
	fmt.Println(os.Stdout)
	fmt.Println("latency_x = ", latency_x)
	fmt.Println("latency_x_1 = ", latency_x_1)
	fmt.Println("avg_latency = ", avg_latency_x_1)
	fmt.Println("m = ", m)
	fmt.Println("rtt = ", m.rtt)
	fmt.Println("pkt_sent = ", m.pkt_sent)
	fmt.Println("pkt_rcvd = ", m.pkt_rcvd)
	fmt.Println("jitter = ", jitter_x)
}
