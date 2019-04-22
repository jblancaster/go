package main

import (
	"fmt"
	"math"
	"time"
	ping "github.com/sparrc/go-ping"
)

type Message struct {
	rtt time.Duration
	pkt_sent int
	pkt_rcvd int
}

const (
	N = 60.0
)

var	(
	err error
	alpha float64 = 0
	s_x float64 = 0
	t_x float64 = 0
	jitter_x   float64 = 0
	jitter_x_1 float64 = 0
	latency_x   float64 = 0
	latency_x_1 float64 = 0
	avg_latency_x   float64 = 0
	avg_latency_x_1 float64 = 0
	avg_jitter_x    float64 = 0
	avg_jitter_x_1  float64 = 0
	avg_variance_latency_x   float64 = 0
	avg_variance_latency_x_1 float64 = 0
	avg_variance_jitter_x    float64 = 0
	avg_variance_jitter_x_1  float64 = 0
	packet_counter int = 0
	recv_counter int = 0
	new_packet_loss int = 0
	s1 int = 0
	s2 int = 0
	pl_avg1 float64 = 0
	pl_avg2 float64 = 0
	effective_latency float64 = 0
	pl_percent int = 0
	r_value float64 = 0
)

func abs_time_diff(t1, t2 float64) float64 {
	v := t1 - t2
	if ( v < 0 ) {
		return v * -1
	} else {
		return v
	}
}

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
	if err != nil {
		return
	}

	// Initialize variables
	alpha = (2.0 / (N+1.0))
	latency_x_1 = float64(float64(int64(m.rtt))/float64(time.Second))
	avg_latency_x_1 = latency_x_1

	for i:=0; i<10; i++ {
		// LOOP - Get new data point
		m, err = ping_it("www.google.com")
		fmt.Println("rtt = ", m.rtt)
		latency_x = float64(float64(int64(m.rtt))/float64(time.Second))
		packet_counter += m.pkt_sent
		recv_counter += m.pkt_rcvd

		// Calculate Latency, Jitter average, variance
		jitter_x = math.Abs(latency_x - latency_x_1)

		// Moving average of latency
		s_x = latency_x - avg_latency_x_1
		avg_latency_x = avg_latency_x_1 + (alpha * s_x)

		// Moving average variance of latency
		avg_variance_latency_x = (1.0 - alpha) * (avg_variance_latency_x_1 + (alpha * s_x * s_x))

		// Moving average of jitter
		t_x = jitter_x - jitter_x_1
		avg_jitter_x = avg_jitter_x_1 + (alpha * t_x)

		// Moving average variance of jitter
		avg_variance_jitter_x = (1.0 - alpha) * (avg_variance_jitter_x_1 + (alpha * t_x * t_x))

		// Write buffer here
		// latency, avg_latency, avg_variance_latency
		// jitter, avg_jitter, avg_variance_jitter
		fmt.Println("alpha = ", alpha)
		fmt.Println("s_x = ", s_x)
		fmt.Println("s_x*s_x = ", s_x*s_x)
		fmt.Println("latency_x = ", latency_x)
		fmt.Println("latency_x_1 = ", latency_x_1)
		fmt.Println("avg_latency = ", avg_latency_x_1)
		fmt.Println("avg_variance_latency_x = ", avg_variance_latency_x)
		fmt.Println("avg_variance_latency_x_1 = ", avg_variance_latency_x_1)
		fmt.Println("avg_variance_jitter_x = ", avg_variance_jitter_x)
		fmt.Println("avg_variance_jitter_x_1 = ", avg_variance_jitter_x_1)
		fmt.Println("pkt_sent = ", m.pkt_sent)
		fmt.Println("pkt_rcvd = ", m.pkt_rcvd)
		fmt.Println("jitter_x = ", jitter_x)
		fmt.Println("jitter_x_1 = ", jitter_x_1)
		fmt.Println("avg_jitter_x = ", avg_jitter_x)
		fmt.Println("avg_jitter_x_1 = ", avg_jitter_x_1)
		fmt.Println("-------------------------------------------------------")

		// Realign variables
		latency_x_1 = latency_x
		avg_latency_x_1 = avg_latency_x
		avg_variance_latency_x_1 = avg_variance_latency_x
		jitter_x_1 = jitter_x
		avg_jitter_x_1 = avg_jitter_x
		avg_variance_jitter_x_1 = avg_variance_jitter_x

		// Packet Loss - On data buffer send
		new_packet_loss = (recv_counter*100) / packet_counter
		s1 = new_packet_loss - pl_percent
		s2 = new_packet_loss - pl_percent
		pl_avg1 = pl_avg1 + (alpha * float64(s1))
		pl_avg2 = pl_avg2 + (alpha * float64(s2))

		pl_avg1 = math.Max(pl_avg1, 0)
		pl_avg1 = math.Min(pl_avg1, 100)
		pl_avg2 = math.Max(pl_avg2, 0)
		pl_avg2 = math.Min(pl_avg2, 100)

		effective_latency = avg_latency_x + (avg_jitter_x * 2.0) + 10.0
		if effective_latency < 160.0 {
			r_value = 93.2 - (effective_latency / 40.0)
		} else {
			r_value = 93.2 - ((effective_latency / 120.0) / 10.0)
		}

		r_value = math.Max(r_value - (pl_avg1 * 2.5), 0.0)

		// Write buffer here
		// packet_counter, recv_counter, pl_avg1, pl_avg2, r_value
		fmt.Println("s1 = ", s1)
		fmt.Println("s2 = ", s2)
		fmt.Println("pl_avg1 = ", pl_avg1)
		fmt.Println("pl_avg2 = ", pl_avg2)
		fmt.Println("effective_latency = ", effective_latency)
		fmt.Println("r_value = ", r_value)
		fmt.Println("packet_counter = ", packet_counter)
		fmt.Println("recv_counter = ", recv_counter)
		fmt.Println("new_packet_loss = ", new_packet_loss)
		fmt.Println("=======================================================")

		pl_percent = new_packet_loss
		//packet_counter = 0
		//recv_counter = 0

		// REPEAT LOOP
		time.Sleep(3 * time.Second)
	}
}
