package main

import (
	"flag"
	"fmt"
//	"encoding/json"
	"math"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"
	ping "github.com/sparrc/go-ping"
	curl "github.com/andelf/go-curl"
)

const (
	default_url string = "www.google.com"
)

var	(
	// Input Args
	help string = "latency <-u URL> <-f Freq/min> <-r Report*min> <-n N> <-n1 N1> <-n2 N2>"
	ping_url string = default_url
	ping_freq int = 4 // Ping per minute
	ping_report int = 1 // Report each x*minute(s)
	ping_n int = 100
	ping_n1 int = 100
	ping_n2 int = 100

	// Algo Parameters
	err error
	alpha float64 = 0
	alpha1 float64 = 0
	alpha2 float64 = 0
	s_x float64 = 0
	t_x float64 = 0
	jitter_x   float64 = 0
	jitter_x1 float64 = 0
	latency_x   float64 = 0
	latency_x1 float64 = 0
	avg_latency_x   float64 = 0
	avg_latency_x1 float64 = 0
	avg_jitter_x    float64 = 0
	avg_jitter_x1  float64 = 0
	avg_variance_latency_x   float64 = 0
	avg_variance_latency_x1 float64 = 0
	avg_variance_jitter_x    float64 = 0
	avg_variance_jitter_x1  float64 = 0
	packet_counter int = 0
	recv_counter int = 0

	// Output Report
	new_packet_loss float64 = 0
	s1 float64 = 0
	s2 float64 = 0
	pl_avg1 float64 = 0
	pl_avg2 float64 = 0
	effective_latency float64 = 0
	pl_percent float64 = 0
	r_value float64 = 0
	latency []time.Duration
	avg_latency []time.Duration
	avg_variance_latency []time.Duration
	jitter []time.Duration
	avg_jitter []time.Duration
	avg_variance_jitter []time.Duration
	packet_counter_out []int
	recv_counter_out []int
	pl_avg1_out []string
	pl_avg2_out []string
	r_value_out []string
	N float64 = 100.0
	N1 float64 = 100.0
	N2 float64 = 100.0

	// Output paramters
	post_url string = "do-lp.komodowifi.com/telem/dt001"

	// Control Signals
	quitting bool = false
)

type Message struct {
	rtt time.Duration
	pkt_sent int
	pkt_rcvd int
}

// JSON response structure
type Latency struct {
	devUUID string
}

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

func calculate_latency(iterations, sleep_time int) {
	for i:=0; i<iterations; i++ {
		// LOOP - Get new data point
		m, err := ping_it(ping_url)
		if err != nil {
			fmt.Println("Cannot ping host")
			panic(err)
			return
		}

		latency_x = float64(float64(int64(m.rtt))/float64(time.Second))
		packet_counter += m.pkt_sent
		recv_counter += m.pkt_rcvd

		// Calculate Latency, Jitter average, variance
		jitter_x = math.Abs(latency_x - latency_x1)

		// Moving average of latency
		s_x = latency_x - avg_latency_x1
		avg_latency_x = avg_latency_x1 + (alpha * s_x)

		// Moving average variance of latency
		avg_variance_latency_x = (1.0 - alpha) * (avg_variance_latency_x1 + (alpha * s_x * s_x))

		// Moving average of jitter
		t_x = jitter_x - jitter_x1
		avg_jitter_x = avg_jitter_x1 + (alpha * t_x)

		// Moving average variance of jitter
		avg_variance_jitter_x = (1.0 - alpha) * (avg_variance_jitter_x1 + (alpha * t_x * t_x))

		// Write buffer here
		// latency, avg_latency, avg_variance_latency
		// jitter, avg_jitter, avg_variance_jitter
		latency = append(latency, time.Duration(latency_x * float64(time.Second)))
		avg_latency = append(avg_latency, time.Duration(avg_latency_x * float64(time.Second)))
		avg_variance_latency = append(avg_variance_latency, time.Duration(avg_variance_latency_x * float64(time.Second)))

		jitter = append(jitter, time.Duration(jitter_x * float64(time.Second)))
		avg_jitter = append(avg_jitter, time.Duration(avg_jitter_x * float64(time.Second)))
		avg_variance_jitter = append(avg_variance_jitter, time.Duration(avg_variance_jitter_x * float64(time.Second)))

		fmt.Println("s_x = ", s_x)
		fmt.Println("s_x*s_x = ", s_x*s_x)
		fmt.Println("latency_x = ", latency_x)
		fmt.Println("avg_latency = ", avg_latency_x1)
		fmt.Println("avg_variance_latency_x = ", avg_variance_latency_x)
		fmt.Println("avg_variance_jitter_x = ", avg_variance_jitter_x)
		fmt.Println("jitter_x = ", jitter_x)
		fmt.Println("avg_jitter_x = ", avg_jitter_x)

		// Realign variables
		latency_x1 = latency_x
		avg_latency_x1 = avg_latency_x
		avg_variance_latency_x1 = avg_variance_latency_x
		jitter_x1 = jitter_x
		avg_jitter_x1 = avg_jitter_x
		avg_variance_jitter_x1 = avg_variance_jitter_x

		// Packet Loss - On data buffer send
		new_packet_loss = float64((recv_counter*100) / packet_counter)
		s1 = new_packet_loss - pl_percent
		s2 = new_packet_loss - pl_percent
		pl_avg1 = pl_avg1 + (alpha1 * float64(s1))
		pl_avg2 = pl_avg2 + (alpha2 * float64(s2))

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
		packet_counter_out = append(packet_counter_out, packet_counter)
		recv_counter_out = append(recv_counter_out, recv_counter)
		pl_avg1_out = append(pl_avg1_out, fmt.Sprintf("%.4f", pl_avg1))
		pl_avg2_out = append(pl_avg2_out, fmt.Sprintf("%.4f", pl_avg2))
		r_value_out = append(r_value_out, fmt.Sprintf("%.4f", r_value))

		fmt.Println("effective_latency = ", effective_latency)
		fmt.Println("r_value = ", r_value)
		fmt.Println("new_packet_loss = ", new_packet_loss)
		fmt.Println("=======================================================")

		pl_percent = new_packet_loss
		//packet_counter = 0
		//recv_counter = 0

		// Break on signal
		if quitting {
			break
		} else {
			// REPEAT LOOP
			time.Sleep(time.Duration(sleep_time) * time.Second)
		}
	}
}

func marshal_json() string {
	// FIXME! - We need to use json/encode here.
	// Output json file
	output :=         fmt.Sprintf("{\n")
	output = output + fmt.Sprintf("  \"devUUID\" : \"12345678901234567890123456789012\",\n")
	output = output + fmt.Sprintf("  \"accountUUID\" : \"12345678901234567890123456789012\",\n")
	output = output + fmt.Sprintf("  \"latency_timestamp\" : \"%d\",\n", time.Now().Unix())
	output = output + fmt.Sprintf("  \"data\" : {\n")
	output = output + fmt.Sprintf("    \"latency\" : \"%v\",\n", latency)
	output = output + fmt.Sprintf("    \"avg_latency\" : \"%v\",\n", avg_latency)
	output = output + fmt.Sprintf("    \"avg_variance_latency\" : \"%v\",\n", avg_variance_latency)
	output = output + fmt.Sprintf("    \"jitter\" : \"%v\",\n", jitter)
	output = output + fmt.Sprintf("    \"avg_jitter\" : \"%v\",\n", avg_jitter)
	output = output + fmt.Sprintf("    \"avg_variance_jitter\" : \"%v\",\n", avg_variance_jitter)
	output = output + fmt.Sprintf("    \"packet_counter\" : \"%v\",\n", packet_counter_out)
	output = output + fmt.Sprintf("    \"recv_counter\" : \"%v\",\n", recv_counter_out)
	output = output + fmt.Sprintf("    \"pl_avg1\" : \"%v\",\n", pl_avg1_out)
	output = output + fmt.Sprintf("    \"pl_avg2\" : \"%v\",\n", pl_avg2_out)
	output = output + fmt.Sprintf("    \"pl_percent\" : \"%v%%\",\n", 100.0 - pl_percent)
	output = output + fmt.Sprintf("    \"r_value\" : \"%v\"\n", r_value_out)
	output = output + fmt.Sprintf("    }\n")
	output = output + fmt.Sprintf("}")
	fmt.Println(output)

	// Clear slices - Note: We're still maintaining iterated values
	latency = latency[:0]
	avg_latency = avg_latency[:0]
	avg_variance_latency = avg_variance_latency[:0]

	jitter = jitter[:0]
	avg_jitter = avg_jitter[:0]
	avg_variance_jitter = avg_variance_jitter[:0]

	packet_counter_out = packet_counter_out[:0]
	recv_counter_out = recv_counter_out[:0]

	pl_avg1_out = pl_avg1_out[:0]
	pl_avg2_out = pl_avg2_out[:0]
	r_value_out = r_value_out[:0]
	return output
}

func curl_results(output string) {
	// Curl results
	var sent = false
	easy := curl.EasyInit()
	defer easy.Cleanup()

	easy.Setopt(curl.OPT_URL, post_url)
	easy.Setopt(curl.OPT_POST, true)
	easy.Setopt(curl.OPT_VERBOSE, true)
	easy.Setopt(curl.OPT_READFUNCTION,
		func(ptr []byte, userdata interface{}) int {
			if !sent {
				sent = true
				ret := copy(ptr, output)
				return ret
			}
			return 0 // sent ok
		})

	// Disable HTTP/1.1 Expect 100
	easy.Setopt(curl.OPT_HTTPHEADER, []string{"Expect:"})

	// Must set
	easy.Setopt(curl.OPT_POSTFIELDSIZE, len(output))

	if err := easy.Perform(); err != nil {
		println("ERROR: ", err.Error())
	}

	time.Sleep(10000) // Wait goroutine
}

// Init
func arg_init() {
	flag.StringVar(&ping_url, "u",  default_url, "URL host to ping against")
	flag.IntVar(&ping_freq,   "f",  ping_freq,   "Number of pings per minute")
	flag.IntVar(&ping_report, "r",  ping_report, "Number of minutes between reports")
	flag.IntVar(&ping_n,      "n",  ping_n,  "N value")
	flag.IntVar(&ping_n1,     "n1", ping_n1, "N1 value")
	flag.IntVar(&ping_n2,     "n2", ping_n2, "N2 value")
	flag.StringVar(&help,     "h",  help,    "This menu")
}

// ping
func main() {
	// Input Arguments
	// latency <-u URL> <-f Freq/min> <-r Report*min> <-n N> <-n1 N1> <-n2 N2>
	arg_init()
	flag.Parse()

	u, err := url.ParseRequestURI(ping_url)
	if u == nil {
		fmt.Println(ping_url, " was invalid. Using default URL. err = ", err)
		ping_url = default_url
	}

	fmt.Println("URL Host is:\t\t",    ping_url)
	fmt.Println("Ping frequecy is\t",  ping_freq,   "per minute")
	fmt.Println("Report every\t\t",    ping_report, "minute(s)")
	fmt.Println("ping_n has value\t",  ping_n)
	fmt.Println("ping_n1 has value\t", ping_n1)
	fmt.Println("ping_n2 has value\t", ping_n2)
	fmt.Println("=======================================================")

	// Initialize variables
	N  = float64(ping_n)
	N1 = float64(ping_n1)
	N2 = float64(ping_n2)
	alpha  = (2.0 /(N+1.0))
	alpha1 = (2.0 /(N1+1.0))
	alpha2 = (2.0 /(N2+1.0))
	sleep_time := int(60/ping_freq)
	iterations := ping_freq * ping_report

	// Set up signals
	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		fmt.Printf("Caught sig: %+v\n", sig)
		fmt.Println("Wait...\n")
		quitting = true
	}()

	// URL, ping_frequency, report_frequency, N, N1, N2, 
	m, err := ping_it(ping_url)
	if err != nil {
		fmt.Println("Cannot ping host")
		return
	}
	// Prime the algorithm
	avg_latency_x1 = latency_x1
	latency_x1 = float64(float64(int64(m.rtt))/float64(time.Second))

	for !quitting {
		// Calculate latency, jitter, stddev, variance, etc.
		calculate_latency(iterations, sleep_time)

		// Marshal the results into a json file and curl it to the host
		output := marshal_json()

		curl_results(output)
	}
	fmt.Println("Done")
}
