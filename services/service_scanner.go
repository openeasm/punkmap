package services

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
	"runtime/pprof"
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
	IP        string `json:"ip"`
	Port      string `json:"port"`
	Open      bool   `json:"open"`
	ErrorMsg  string `json:"error_msg,omitempty"`
	Protocol  string `json:"protocol,omitempty"`
	Service   string `json:"service,omitempty"`
	Banner    string `json:"banner,omitempty"`
	BannerHex []byte `json:"banner_hex,omitempty"`
	Time      int64  `json:"time"`
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
}

func (m *Metrics) LogSuccessRate() {
	log.Printf("LogSuccessRate: %f,has banner rate: %f\n", float64(m.Open)/float64(m.Total), float64(m.HasBanner)/float64(m.Total))
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
	result := &Result{IP: t.ip, Port: t.port, Open: true, Protocol: "tcp"}
	if PortScannersMapping[t.port] != nil {
		for _, scanner := range PortScannersMapping[t.port] {
			conn, err := net.DialTimeout("tcp", t.ip+":"+t.port, time.Duration(t.timeout)*time.Second)
			if err != nil {
				return &Result{IP: t.ip, Port: t.port, Open: false, ErrorMsg: err.Error(), Protocol: "tcp"}
			}
			defer conn.Close()
			service, banner, err := scanner.Scan(conn, t)
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
func (s *Scanner) ScanWorker(inputChan chan string, outputChan chan *Result, wg *sync.WaitGroup) {
	defer wg.Done()
	for ipAddr := range inputChan {
		atomic.AddInt64(&s.Processing, 1)
		atomic.AddInt64(&s.Total, 1)

		if strings.Contains(ipAddr, ":") { //
			// split ip:port
			ip := strings.Split(ipAddr, ":")[0]
			port := strings.Split(ipAddr, ":")[1]
			if s.Debug {
				log.Printf("start scan %s:%s", ip, port)
			}
			// 开始扫描
			startTime := time.Now().UnixMilli()
			task := Task{ip: ip, port: port, timeout: s.Timeout}
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
				log.Printf("finish scan %s:%s, open:%t . time_cost:%d ms", ip, port, result.Open, totalCost)
			}
			atomic.AddInt64(&s.Processing, -1)
			outputChan <- result
		} else {
			for _, port := range strings.Split(s.Ports, ",") {
				task := Task{ip: ipAddr, port: port, timeout: s.Timeout}
				result := s.Scan(task)
				outputChan <- result
			}
		}
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
	var enc = json.NewEncoder(pipe)

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
				if err := enc.Encode(&response); err != nil {
					log.Fatal(err)
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
				jsonData, err := json.Marshal(response)
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
	if s.EnableCpuProfile {
		cpuf, err := os.Create(s.CpuProfileName)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(cpuf)
		defer pprof.StopCPUProfile()
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
		nc, _ := nats.Connect(s.InputNatsURL)
		js, _ := jetstream.New(nc)
		timeoutCtx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		stream, _ := js.CreateStream(timeoutCtx, jetstream.StreamConfig{
			Name:     s.InputNatsJS,
			Subjects: []string{s.InputNatsSubject},
		})
		c, _ := stream.CreateOrUpdateConsumer(timeoutCtx, jetstream.ConsumerConfig{
			Name: s.NatsWorkerName,
		})
		for {
			//msg, err := c.Next()
			msgs, err := c.FetchNoWait(s.ProcessNum * 4)
			if err != nil {
				log.Println(err)
				time.Sleep(10 * time.Second)
				continue
			}
			for {
				if msgs == nil {
					break
				}
				msg := <-msgs.Messages()
				if err != nil {
					break
				}
				if msg == nil {
					break
				}
				inputChan <- string(msg.Data())
				msg.Ack()
			}
			time.Sleep(1 * time.Second)
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
