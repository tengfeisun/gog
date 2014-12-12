package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var hostport string

var receivedNum int32
var startTime time.Time
var elaspedTime time.Duration

var vmu sync.Mutex
var aViewCnt = make(map[string]int)
var pViewCnt = make(map[string]int)
var times = make(map[string]int)

var testSerf bool

var htmldir string

var messageArray []string

func handleStart(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	startTime = time.Now()
	elaspedTime = 0
	receivedNum = 0
	times = make(map[string]int)
	aViewCnt = make(map[string]int)
	pViewCnt = make(map[string]int)
}

func handleReceived(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	msg := r.Form.Get("color")
	//reqstart := time.Now()
	num := atomic.AddInt32(&receivedNum, 1)
	if num == 1 {
		startTime = time.Now()
	}
	elaspedTime = time.Now().Sub(startTime)
	if !testSerf {
		if num%100 == 0 {
			times[fmt.Sprintf("%d", num)] = int(elaspedTime.Nanoseconds() / 1000000)
		}
	}

	messageArray[num-1] = msg
	//fmt.Fprintf(w, "hello %d, time: %v, req time: %v", num, elaspedTime, time.Now().Sub(reqstart))
}

func handleQuery(w http.ResponseWriter, r *http.Request) {
	if !testSerf {
		for i := range times {
			fmt.Fprintf(w, "Received: %s, time: %dms\n", i, times[i])
		}
	}
	fmt.Fprintf(w, "total received: %d, elasped time: %v\n", receivedNum, elaspedTime)

	vmu.Lock()
	defer vmu.Unlock()

	avg, std := computeView(aViewCnt)
	fmt.Fprintf(w, "Aview avg: %v, std: %v\n", avg, std)
	avg, std = computeView(pViewCnt)
	fmt.Fprintf(w, "Pview avg: %v, std: %v\n", avg, std)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	indexPath := path.Join(htmldir, "/index.html")

	indexFile, err := os.Open(indexPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = io.Copy(w, indexFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleJson(w http.ResponseWriter, r *http.Request) {
	b, err := json.Marshal(messageArray)
	if err != nil {
		fmt.Println("error handle json", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, string(b))
}

func handleDemoJson(w http.ResponseWriter, r *http.Request) {
	times["0"] = 0
	times[fmt.Sprintf("%d", int(receivedNum))] = int(elaspedTime.Nanoseconds() / 1000000)

	b, err := json.Marshal(times)
	if err != nil {
		fmt.Println("error handle json demo", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, string(b))
}

func handleView(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	// "aviewCnt:pviewCnt:id"
	view := strings.SplitAfterN(string(b), ":", 3)

	av, err := strconv.Atoi(view[0][:len(view[0])-1])
	if err != nil {
		fmt.Println(err)
		return
	}
	pv, err := strconv.Atoi(view[1][:len(view[1])-1])
	if err != nil {
		fmt.Println(err)
		return
	}

	vmu.Lock()
	defer vmu.Unlock()
	aViewCnt[view[2]] = av
	pViewCnt[view[2]] = pv
}

// Return average, std
func computeView(view map[string]int) (float64, float64) {
	node := 0
	total := 0
	for _, v := range view {
		node++
		total = total + v
	}
	avg := float64(total) / float64(node)

	varstd := 0.0
	for _, v := range view {
		varstd = varstd + (float64(v)-avg)*(float64(v)-avg)
	}
	varstd = varstd / float64(node)
	std := math.Sqrt(varstd)
	return avg, std
}

func init() {
	flag.StringVar(&hostport, "hostport", ":11000", "The server's address")
	flag.BoolVar(&testSerf, "testserf", false, "If testing serf")
	flag.StringVar(&htmldir, "htmldir", "", "The htmldir")
}

func main() {
	flag.Parse()

	messageArray = make([]string, 900)
	for i := range messageArray {
		messageArray[i] = "white"
	}

	fmt.Println("Start server...")
	http.HandleFunc("/index", handleIndex)
	http.HandleFunc("/start", handleStart)
	http.HandleFunc("/received", handleReceived)
	http.HandleFunc("/query", handleQuery)
	http.HandleFunc("/view", handleView)
	http.HandleFunc("/json", handleJson)
	http.HandleFunc("/demojson", handleDemoJson)

	if err := http.ListenAndServe(hostport, nil); err != nil {
		fmt.Println(err)
	}
}
