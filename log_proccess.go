package main

import (
	"strings"
	"fmt"
	"time"
	"os"
	"bufio"
	"io"
	"regexp"
	"log"
	"strconv"
	"net/url"
	"github.com/influxdata/influxdb/client/v2"
	"flag"
	"net/http"
	"encoding/json"
)

type Reader interface {
	Read(rc chan []byte)
}

type Writer interface {
	Write(wc chan *Message)
}

type Message struct {
	TimeLocal time.Time
	BytesSent int
	Path,Method,Scheme,Status string
	UpstreamTime,RequestTime float64
}

// 系统状态监控
type SystemInfo struct {
	HandleLine int `json:"handleLine"`
	Tps float64 `json:"tps"`
	ReadChanLen int `json:"readChanLen"`
	WriteChanLen int `json:"writeChanLen"`
	RunTime string `json:"runTime"`
	ErrNum int `json:"errNum"`
}

const (
	TypeHandleLine = 0
	TypeErrNum = 1
)

var TypeMonitorChan = make(chan int, 200)


type Monitor struct {
	startTime time.Time
	data SystemInfo
	tpsSli []int
}

func (m *Monitor) start(lp *LogProcess) {


	go func() {
		for n := range TypeMonitorChan {
			switch n {
			case TypeErrNum:
				m.data.ErrNum += 1
			case TypeHandleLine:
				m.data.HandleLine += 1
			}
		}
	}()


	ticker := time.NewTicker(time.Second * 5)
	go func() {
		for {
			<-ticker.C
			m.tpsSli = append(m.tpsSli, m.data.HandleLine)
			if len(m.tpsSli) > 2 {
				m.tpsSli = m.tpsSli[1:]
			}
		}
	}()


	http.HandleFunc("/monitor", func(writer http.ResponseWriter, request *http.Request) {
		m.data.RunTime = time.Now().Sub(m.startTime).String()
		m.data.ReadChanLen = len(lp.rc)
		m.data.WriteChanLen = len(lp.wc)

		if len(m.tpsSli) >= 2 {
			m.data.Tps = float64(m.tpsSli[1]-m.tpsSli[0]) / 5
		}
		

		ret, _ := json.MarshalIndent(m.data,"","\t")

		io.WriteString(writer,string(ret))
	})

	http.ListenAndServe(":9193",nil)
}


type LogProcess struct {
	rc chan []byte
	wc chan *Message
	read Reader
	write Writer
}

type ReadFromFile struct {
	path string // 读取文件的路径
}

type WriteToInfluxDB struct {
	influxDBDsn string // influx data source
}

func (r *ReadFromFile) Read(rc chan []byte) {
	// 读取模块
	// 打开文件

	f, err := os.Open(r.path)
	if err != nil {
		panic(fmt.Sprintf("open file error:%s", err.Error()))
	}

	// 从文件末尾开始逐行读取文件内容
	f.Seek(0,2)

	rd := bufio.NewReader(f)

	for {
		line, err := rd.ReadBytes('\n')
		if err == io.EOF {
			time.Sleep(500 * time.Millisecond)
			continue
		} else if err != nil {
			panic(fmt.Sprintf("ReadBytes error:%s", err.Error()))
		}
		TypeMonitorChan <- TypeHandleLine
		rc <- line[:len(line)-1]
	}

}

func (w *WriteToInfluxDB) Write(wc chan *Message) {
	// 写入模块

	infSli := strings.Split(w.influxDBDsn, "@")

	// Create a new HTTPClient
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     infSli[0],
		Username: infSli[1],
		Password: infSli[2],
	})
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	// Create a new point batch
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  infSli[3],
		Precision: infSli[4],
	})
	if err != nil {
		log.Fatal(err)
	}

	for v := range wc {
		// Create a point and add to batch
		tags := map[string]string{"Path": v.Path, "Method": v.Method, "Scheme": v.Scheme, "Status": v.Status}
		fields := map[string]interface{}{
			"UpstreamTime":   v.UpstreamTime,
			"RequestTime": v.RequestTime,
			"BytesSent":   v.BytesSent,
		}

		pt, err := client.NewPoint("nginx_log", tags, fields, time.Now())
		if err != nil {
			log.Fatal(err)
		}
		bp.AddPoint(pt)

		// Write the batch
		if err := c.Write(bp); err != nil {
			log.Fatal(err)
		}

		// Close client resources
		if err := c.Close(); err != nil {
			log.Fatal(err)
		}

		log.Println("write success")
	}

}

func (l *LogProcess) Process(){
	// 解析模块

	/**
	[04/Mar/2018:13:49:53 +0000] http "GET /foo?query=t HTTP/1.0" 200 2133 "-"
	"KeepAliveClient" "-" 1.005 1.854
	*/

	r := regexp.MustCompile(`\[([^\]]+)\]\s+(.*?)\s+\"(.*?)\"\s+(\d{3})\s+(\d+)\s+\"([^"]+)\"\s+\"(.*?)\"\s+\"([\d\.-]+)\"\s+([\d\.-]+)\s+([\d\.-]+)`)

	loc, _ := time.LoadLocation("Asia/Shanghai")
	for v := range l.rc {
		ret := r.FindStringSubmatch(string(v))
		if len(ret) != 11 {
			TypeMonitorChan <- TypeErrNum
			log.Println("FindStringSubmatch fail:", string(v))
			continue
		}

		message := &Message{}
		t, err := time.ParseInLocation("09/Jan/2006:15:04:05 +0000",ret[1],loc)
		if err != nil {
			TypeMonitorChan <- TypeErrNum
			log.Println("ParseInLocation fail:",err.Error(), ret[1])
			continue
		}

		message.TimeLocal = t

		byteSent, _ := strconv.Atoi(ret[5])
		message.BytesSent = byteSent

		// GET /foo?query=t HTTP/1.0
		reqSli := strings.Split(ret[3]," ")
		if len(reqSli) != 3 {
			TypeMonitorChan <- TypeErrNum
			log.Println("strings.Split fail",ret[3])
			continue
		}

		message.Method = reqSli[0]

		u, err := url.Parse(reqSli[1])
		if err != nil {
			TypeMonitorChan <- TypeErrNum
			log.Println("url parse fail:", err.Error())
			continue
		}

		message.Path = u.Path

		message.Scheme = ret[2]
		message.Status = ret[4]

		upstramTime , _ :=strconv.ParseFloat(ret[9], 64)
		requestTime , _ :=strconv.ParseFloat(ret[10], 64)
		message.UpstreamTime = upstramTime
		message.RequestTime = requestTime


		l.wc <- message
	}
}


func main() {

	var path, influxDsn string
	flag.StringVar(&path, "path", "./access.log", "read file path")
	flag.StringVar(&influxDsn, "influxDsn", "http://127.0.0.1:8086@myself@myselfpass@myself@s", "influx data source")

	flag.Parse()

	r := &ReadFromFile{
		path:path,
	}

	w := &WriteToInfluxDB{
		influxDBDsn:influxDsn,
	}


	lp := &LogProcess{
		rc:make(chan []byte),
		wc:make(chan *Message),
		read:r,
		write:w,
	}

	go lp.read.Read(lp.rc)
	go lp.Process()
	go lp.write.Write(lp.wc)

	m := Monitor{
		startTime:time.Now(),
		data:SystemInfo{},
	}

	m.start(lp)


}
