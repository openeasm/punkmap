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
	"strings"
	"sync"
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

type Scanner struct {
	// below is for stdin/stdout
	InputFile  string `short:"i" long:"input" description:"input file"  default:"-"`
	OutputFile string `short:"o" long:"output" description:"output file"  default:"-"`
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

	OutputHex   bool `long:"output-hex" description:"output base64"`
	OutputClose bool `long:"output-close" description:"output closed ports"`
}

func (s *Scanner) Scan(t Task) (r *Result) {
	// dial ip:port
	conn, err := net.DialTimeout("tcp", t.ip+":"+t.port, time.Duration(t.timeout)*time.Second)
	if err != nil {
		return &Result{IP: t.ip, Port: t.port, Open: false, ErrorMsg: err.Error(), Protocol: "tcp"}
	}
	defer conn.Close()
	// read data from conn

	result := &Result{IP: t.ip, Port: t.port, Open: true, Protocol: "tcp"}
	if PortScannersMapping[t.port] != nil {
		for _, scanner := range PortScannersMapping[t.port] {
			service, banner, err := scanner.Scan(conn, t)
			if err != nil {
				result.ErrorMsg = err.Error()
			}
			result.Service = service
			result.BannerHex = banner
		}
	}

	return result
}
func (s *Scanner) ScanWorker(inputChan chan string, outputChan chan *Result, wg *sync.WaitGroup) {
	defer wg.Done()
	for ipAddr := range inputChan {
		if strings.Contains(ipAddr, ":") { //
			// split ip:port
			ip := strings.Split(ipAddr, ":")[0]
			port := strings.Split(ipAddr, ":")[1]
			if s.Debug {
				log.Printf("start scan %s:%s", ip, port)
			}
			task := Task{ip: ip, port: port, timeout: s.Timeout}
			result := s.Scan(task)
			result.Time = time.Now().Unix()
			if s.Debug {
				log.Printf("scan %s:%s, open:%t ", ip, port, result.Open)
			}
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

func (s *Scanner) Start() {
	inputChan := make(chan string, s.ProcessNum*4)
	outputChan := make(chan *Result, s.ProcessNum*4)

	// start workers
	outputWg := sync.WaitGroup{}
	scanWg := sync.WaitGroup{}
	if len(s.OutputNatsURL) > 0 {
		for i := 0; i < s.NatsWriterNum; i++ {
			go s.NatsWriteWorker(outputChan, &outputWg)
		}

	} else {
		go s.WriteWorker(outputChan, &outputWg)
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
				continue
			}
			for {
				msg := <-msgs.Messages()
				if err != nil {
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
