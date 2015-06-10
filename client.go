package http

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"regexp"
	"strconv"
	"time"
)

type Client struct {
	Socket *net.Conn
	Host   string
	Port   int
	Https  bool
}

type Result struct {
	TimeSending           float64
	TimeReadingFirstBytes float64
	TimeReadingTotal      float64
	TimeTotal             float64
}

func (c *Client) Open(host string, port int, https bool) {
	socket, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		panic(err)
	}

	c.Socket = &socket
	c.Host = host
	c.Port = port
	c.Https = https
}

func (c *Client) Get(route string) (Result, int) {
	if c.Socket == nil {
		panic("Socket not opened")
	}

	// Prepare HTTP request.
	var url string
	if c.Https {
		url = fmt.Sprintf("https://%s%s", c.Host, route)
	} else {
		url = fmt.Sprintf("http://%s%s", c.Host, route)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}
	dump, err := httputil.DumpRequest(req, false)
	if err != nil {
		panic(err)
	}

	// Write/Read HTTP request/response.
	results := Result{}
	startSending := time.Now()
	fmt.Fprintf((*c.Socket), string(dump))
	results.TimeSending = time.Since(startSending).Seconds() * 1000
	startReading := time.Now()
	reader := bufio.NewReader((*c.Socket))
	status, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}
	results.TimeReadingFirstBytes = time.Since(startReading).Seconds() * 1000
	for ; err != nil; _, err = reader.ReadByte() {
		// Read whole response.
	}
	results.TimeReadingTotal = time.Since(startReading).Seconds() * 1000
	results.TimeTotal = time.Since(startSending).Seconds() * 1000

	// Parse code.
	re := regexp.MustCompile("HTTP/1.[0-1] ([0-9]{3}).*")
	submatches := re.FindStringSubmatch(status)
	if len(submatches) == 0 {
		panic("Can't find HTTP response code")
	}
	code, err := strconv.Atoi(submatches[1])
	if err != nil {
		panic(err)
	}

	return results, code
}

func (c *Client) Close() {
	(*c.Socket).Close()
}
