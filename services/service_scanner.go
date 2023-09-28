package services

import (
	"github.com/pkg/profile"
)

import (
	"bufio"
	"context"
	"easm_punkmap/common"
	"encoding/json"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"log"

	"net"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Task struct {
	ip      string
	port    string
	timeout int
}

func (task Task) ToHttpHost() string {
	if task.port == "80" || task.port == "443" {
		return task.ip
	} else {
		return task.ip + ":" + task.port
	}
}

type Result struct {
	IP          string                 `json:"ip"`
	ConnIP      string                 `json:"conn_ip,omitempty"`
	Port        string                 `json:"port"`
	Open        bool                   `json:"open"`
	ErrorMsg    string                 `json:"error_msg,omitempty"`
	Protocol    string                 `json:"protocol,omitempty"`
	Service     string                 `json:"service,omitempty"`
	ServiceMeta map[string]interface{} `json:"-"`
	Banner      string                 `json:"banner,omitempty"`
	BannerHex   []byte                 `json:"banner_hex,omitempty"`
	Time        int64                  `json:"time"`
}

func (b *Result) ToJson() ([]byte, error) {
	if b.ServiceMeta != nil {
		var inInterface map[string]interface{}
		inrec, err := json.Marshal(b)
		if err != nil {
			log.Fatal(err)
		}
		json.Unmarshal(inrec, &inInterface)
		// merge the two maps
		for k, v := range b.ServiceMeta {
			//fmt.Println("k:", k, "v:", v)
			inInterface[k] = v
		}
		return json.Marshal(inInterface)

		//data:= json.Marshal(b.ServiceMeta)

	} else {
		return json.Marshal(b)
	}
}

// Metrics 统计指标
type Metrics struct {
	Total      int64 `json:"total"`
	Open       int64 `json:"open"`
	Close      int64 `json:"close"`
	Processing int64 `json:"processing"`
	MaxTime    int64 `json:"max_time"`
	MinTime    int64 `json:"min_time"`
	AvgTime    int64 `json:"avg_time"`
	HasBanner  int64 `json:"has_banner"`
	NoBanner   int64 `json:"no_banner"`
	StartTime  int64 `json:"start_time"`
}

func (m *Metrics) LogSuccessRate() {
	if m.Open != 0 && m.Total != 0 {
		log.Printf("total -> open rate: %f,open -> has banner rate: %f, total -> has banenr rate: %f\n, process speed: %f/s",
			float64(m.Open)/float64(m.Total), float64(m.HasBanner)/float64(m.Open), float64(m.HasBanner)/float64(m.Total), float64(m.Total)/float64(time.Now().Unix()-m.StartTime))
	} else {
	}
}

type Scanner struct {
	// below is for stdin/stdout
	Metrics
	PrintMetricsInterval int    `long:"print-metrics-interval" description:"print metrics interval"  default:"10"`
	InputFile            string `short:"i" long:"input" description:"input file"  default:"-"`
	OutputFile           string `short:"o" long:"output" description:"output file"  default:"-"`
	// below is for nats
	InputNatsURL     string `long:"input-nats" description:"ues nats as input"  default:""`
	InputNatsJS      string `long:"input-nats-js" description:"the jetstream name in nats"  default:""`
	InputNatsSubject string `long:"input-nats-subject" description:"the topic name in nats"  default:""`
	InputNatsName    string `long:"input-nats-name" description:"the worker name in nats"  default:""`

	OutputNatsURL     string `long:"output-nats" description:"ues nats as output"  default:""`
	NatsWriterNum     int    `long:"nats-writer-num" description:"the number of nats writer"  default:"10"`
	OutNatsJS         string `long:"output-nats-js" description:"the output jetstream name in nats"  default:""`
	OutputNatsSubject string `long:"output-nats-subject" description:"the output topic name in nats"  default:""`
	OutputNatsGzip    bool   `long:"output-gzip" description:"use gzip to compress output"`

	NatsWorkerName string `long:"nats-worker-name" description:"the worker name in nats"  default:"punkmap-scanner"`
	// below is for scan
	ProcessNum int    `short:"p" long:"process" description:"process number"  default:"10"`
	Timeout    int    `short:"t" long:"timeout" description:"timeout"  default:"3"`
	Retries    int    `short:"r" long:"retries" description:"retries" default:"1"`
	Ports      string `long:"ports" description:"ports to scan" default:"22,9876"`

	Debug bool `short:"d" long:"debug" description:"debug" `

	OutputHex        bool   `long:"output-hex" description:"output base64"`
	OutputClose      bool   `long:"output-close" description:"output closed ports"`
	EnableCpuProfile bool   `long:"enable-cpu-profile" description:"enable cpu profile"`
	CpuProfileName   string `long:"cpu-profile-name" description:"cpu profilename" default:"punkmap_cpu.prof"`
}

func (s *Scanner) ScanWithGlobalTimeout(t Task) (r *Result) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.Timeout)*time.Second*2)
	defer cancel()
	ch := make(chan struct{})
	go func() {
		r = s.Scan(Task{ip: t.ip, port: t.port, timeout: s.Timeout})
		ch <- struct{}{}
	}()
	select {
	case <-ch:
		return
	case <-ctx.Done():
		r = &Result{IP: t.ip, Port: t.port, Open: false, ErrorMsg: "global-timeout", Protocol: "tcp"}
		return
	}
}
func (s *Scanner) Scan(t Task) (r *Result) {
	result := &Result{IP: t.ip, Port: t.port, Open: true, Protocol: "tcp", Time: time.Now().Unix(), ServiceMeta: map[string]interface{}{}}
	if PortScannersMapping[t.port] != nil {
		for _, scanner := range PortScannersMapping[t.port] {
			conn, err := net.DialTimeout("tcp", t.ip+":"+t.port, time.Duration(t.timeout)*time.Second)
			if err != nil {
				return &Result{IP: t.ip, Port: t.port, Open: false, ErrorMsg: err.Error(), Protocol: "tcp"}
			}
			if conn.RemoteAddr().String() != t.ip+":"+t.port {
				connIPPort := conn.RemoteAddr().String()
				result.ConnIP = strings.Split(connIPPort, ":")[0]
			}
			defer conn.Close()
			conn.SetReadDeadline(time.Now().Add(time.Duration(t.timeout) * time.Second))
			service, banner, err := scanner.Scan(conn, t, result)
			if service == "HTTP" || service == "HTTPS" {
				parsedResponse := common.HTTPParser(banner)
				if parsedResponse != nil {
					for k, v := range parsedResponse {
						result.ServiceMeta[k] = v
					}
				}
			}
			if err != nil {
				result.ErrorMsg = err.Error()
			}
			result.Service = service
			result.BannerHex = banner
			if banner != nil {
				break
			}
		}
	}

	return result
}
func (s *Scanner) Task2Result(task Task) *Result {
	startTime := time.Now().UnixMilli()
	if s.Debug {
		log.Printf("start scan %s:%s", task.ip, task.port)
	}
	result := s.ScanWithGlobalTimeout(task)
	endTime := time.Now().UnixMilli()
	if result.Open {
		atomic.AddInt64(&s.Open, 1)
	} else {
		atomic.AddInt64(&s.Close, 1)
	}
	if result.BannerHex != nil {
		atomic.AddInt64(&s.HasBanner, 1)
	} else {
		atomic.AddInt64(&s.NoBanner, 1)
	}
	totalCost := endTime - startTime
	if s.Debug {
		log.Printf("finish scan %s:%s, open:%t . time_cost:%d ms", task.ip, task.port, result.Open, totalCost)
	}
	return result
}
func (s *Scanner) ScanWorker(inputChan chan string, outputChan chan *Result, wg *sync.WaitGroup) {
	defer wg.Done()
	for ipAddr := range inputChan {
		atomic.AddInt64(&s.Processing, 1)
		atomic.AddInt64(&s.Total, 1)
		var tasks []Task
		if strings.Contains(ipAddr, ":") {
			// split ip:port
			ip := strings.Split(ipAddr, ":")[0]
			port := strings.Split(ipAddr, ":")[1]
			tasks = append(tasks, Task{ip: ip, port: port, timeout: s.Timeout})
		} else if strings.Contains(s.Ports, ",") {
			// pre scan mode. ports is split by ,
			for _, port := range strings.Split(s.Ports, ",") {
				tasks = append(tasks, Task{ip: ipAddr, port: port, timeout: s.Timeout})
			}
		}
		for _, task := range tasks {
			result := s.Task2Result(task)
			outputChan <- result
		}
		atomic.AddInt64(&s.Processing, -1)
	}
}
func (s *Scanner) WriteWorker(output chan *Result, outputWg *sync.WaitGroup) {
	defer outputWg.Done()
	var pipe *os.File
	var err error
	if s.OutputFile == "-" {
		pipe = os.Stdout
	} else {
		pipe, err = os.OpenFile(s.OutputFile, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			panic(err)
		}
	}

	for {
		select {
		case response, ok := <-output:
			if ok {
				if !s.OutputHex {
					response.Banner = string(response.BannerHex)
					response.BannerHex = nil
				}
				if !response.Open && !s.OutputClose {
					continue
				}
				data, err := response.ToJson()
				if err != nil {
					pipe.Close()
				} else {
					_, err = pipe.Write(data)
					if err != nil {
						pipe.Close()
					} else {
						_, err = pipe.Write([]byte("\n"))
						if err != nil {
							pipe.Close()
						}
					}
				}

			} else {
				return
			}
		}
	}
}
func (s *Scanner) NatsWriteWorker(output chan *Result, outputWg *sync.WaitGroup) {
	defer outputWg.Done()
	nc, _ := nats.Connect(s.OutputNatsURL)
	js, _ := jetstream.New(nc)
	timeoutCtx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	_, _ = js.CreateStream(timeoutCtx, jetstream.StreamConfig{
		Name:     s.OutNatsJS,
		Subjects: []string{s.OutputNatsSubject},
	})

	for {
		select {
		case response, ok := <-output:
			if ok {
				if !s.OutputHex {
					response.Banner = string(response.BannerHex)
					response.BannerHex = nil
				}
				if !response.Open && !s.OutputClose {
					continue
				}
				jsonData, err := response.ToJson()
				if err != nil {
					log.Println(err)
				}
				if s.OutputNatsGzip {
					jsonData = common.MustGzipEncode(jsonData)
				}
				err = nc.Publish(s.OutputNatsSubject, jsonData)
				if err != nil {
					log.Println(err)
				}
			} else {
				return
			}
		}
	}
}
func (s *Scanner) PrintMatrix() {
	for {

		time.Sleep(time.Duration(s.PrintMetricsInterval) * time.Second)
		processing := atomic.LoadInt64(&s.Metrics.Processing)
		total := atomic.LoadInt64(&s.Metrics.Total)
		open := atomic.LoadInt64(&s.Metrics.Open)
		close := atomic.LoadInt64(&s.Metrics.Close)
		hasBanner := atomic.LoadInt64(&s.Metrics.HasBanner)
		noBanner := atomic.LoadInt64(&s.Metrics.NoBanner)
		log.Printf("processing:%d, total:%d, open:%d, close:%d, hasBanner:%d, noBanner:%d", processing, total, open, close, hasBanner, noBanner)
		s.Metrics.LogSuccessRate()
	}
}

func (s *Scanner) Start() {
	s.Metrics.StartTime = time.Now().Unix()
	if s.EnableCpuProfile {
		defer profile.Start(profile.CPUProfile, profile.ProfilePath(".")).Stop()
	}
	if s.PrintMetricsInterval > 0 {
		go s.PrintMatrix()
	}
	inputChan := make(chan string, s.ProcessNum*4)
	outputChan := make(chan *Result, s.ProcessNum*4)

	// start workers
	outputWg := sync.WaitGroup{}
	scanWg := sync.WaitGroup{}
	if len(s.OutputNatsURL) > 0 {
		for i := 0; i < s.NatsWriterNum; i++ {
			go s.NatsWriteWorker(outputChan, &outputWg)
			outputWg.Add(1)
		}

	} else {
		go s.WriteWorker(outputChan, &outputWg)
		outputWg.Add(1)
	}

	for i := 0; i < s.ProcessNum; i++ {
		go s.ScanWorker(inputChan, outputChan, &scanWg)
		scanWg.Add(1)
	}
	if len(s.InputNatsURL) > 0 {
		nc, err := nats.Connect(s.InputNatsURL)
		if err != nil {
			panic(err)
		}
		if s.Debug {
			log.Printf("connect to nats %s success", s.InputNatsURL)
		}
		js, err := jetstream.New(nc)
		if err != nil {
			panic(err)
		}
		if s.Debug {
			log.Printf("connect to jetstream success")
		}
		timeoutCtx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		stream, _ := js.CreateStream(timeoutCtx, jetstream.StreamConfig{
			Name:     s.InputNatsJS,
			Subjects: []string{s.InputNatsSubject},
		})

		for {
			c, err := stream.CreateOrUpdateConsumer(timeoutCtx, jetstream.ConsumerConfig{
				Name: s.NatsWorkerName,
			})
			if err != nil {
				time.Sleep(1 * time.Second)
				log.Println("create consumer error:", err)
			}
			//msg, err := c.Next()
			var cxt, _ = c.Messages()

			for {
				msg, err := cxt.Next()
				if err != nil {
					break
				}
				if msg == nil {
					break
				}
				inputChan <- string(msg.Data())
				msg.Ack()
			}

		}

	} else {
		// read input file
		var scanner *os.File
		if s.InputFile == "-" {
			scanner = os.Stdin
		} else {
			var err error
			scanner, err = os.Open(s.InputFile)
			if err != nil {
				panic(err)
			}
		}
		// read input file and put to inputChan
		f := bufio.NewScanner(scanner)
		for f.Scan() {
			data := f.Text()
			if data == "" {
				continue
			}
			inputChan <- data
		}
	}

	close(inputChan)
	scanWg.Wait()
	close(outputChan)
	outputWg.Wait()
}
