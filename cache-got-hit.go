package main

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	cli "github.com/caiguanhao/cache-got-hit/vendor/cli-cf1f63a7274872768d4037305d572b70b1199397"
	"github.com/caiguanhao/cache-got-hit/vendor/redigo-4ed1111375cbeb698249ffe48dd463e9b0a63a7a/redis"
)

var (
	redisConn redis.Conn

	noHeaders bool
)

func debug(a ...interface{}) {
	fmt.Fprintln(os.Stderr, a...)
}

func fail(a ...interface{}) {
	debug(a...)
	os.Exit(1)
}

func process(conn net.Conn) {
	writer := bufio.NewWriter(conn)
	defer func() {
		writer.WriteString("END\r\n")
		writer.Flush()
		conn.Close()
	}()

	reader := bufio.NewReader(conn)
	line, _, err := reader.ReadLine()
	if err != nil {
		debug("readline error:", err)
		return
	}

	arr := strings.Fields(string(line))
	if len(arr) < 2 || arr[0] != "get" {
		return
	}

	key := arr[1]
	var etag string
	if len(arr) > 2 {
		etag = arr[2]
	}
	content, err := redis.Bytes(redisConn.Do("GET", key))
	if err != nil {
		if err == redis.ErrNil {
			debug("not found in redis:", key)
		} else {
			debug("redis failed:", err)
		}
		return
	}

	actual := fmt.Sprintf("W/\"%x\"", md5.Sum(content))

	status := 200
	if etag == actual {
		content = nil
		status = 304
	}
	var headers string
	if !noHeaders {
		headers = fmt.Sprintf("-----CACHE GOT HIT HEADERS INCLUDED-----\r\n"+
			"Content-Type: application/json; charset=utf-8\r\n"+
			"Cache-Got-Hit-Status: %d\r\n"+
			"X-Frame-Options: SAMEORIGIN\r\n"+
			"X-XSS-Protection: 1; mode=block\r\n"+
			"X-Content-Type-Options: nosniff\r\n"+
			"ETag: %s\r\n"+
			"\r\n", status, actual)
	}
	writer.WriteString(fmt.Sprintf("VALUE %s 0 %d\r\n", key, len(headers)+len(content)))
	writer.WriteString(headers)
	writer.Write(content)
	writer.WriteString("\r\n")
	debug("cache got hit:", key)
}

func main() {
	app := cli.NewApp()
	app.Name = "cache-got-hit"
	app.Usage = "Let Nginx talk to Redis cache server with modified memcached protocol."
	app.ArgsUsage = " "
	app.HideHelp = true
	app.HideVersion = true
	app.Version = ""
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "bind, b",
			Value: "127.0.0.1:8080",
			Usage: "address and port number to bind to",
		},
		cli.StringFlag{
			Name:  "redis, r",
			Value: "127.0.0.1:6379",
			Usage: "address and port number of redis server",
		},
		cli.IntFlag{
			Name:  "number, n",
			Value: 0,
			Usage: "redis database number",
		},
		cli.IntFlag{
			Name:  "connect-timeout",
			Value: 2,
			Usage: "redis server connection timeout in seconds",
		},
		cli.BoolFlag{
			Name:        "no-headers",
			Usage:       "do not send headers",
			Destination: &noHeaders,
		},
	}
	app.Action = func(c *cli.Context) {
		var redisErr error
		redisConn, redisErr = redis.Dial(
			"tcp",
			c.String("redis"),
			redis.DialDatabase(c.Int("number")),
			redis.DialConnectTimeout(time.Second*time.Duration(c.Int("connect-timeout"))),
		)
		if redisErr != nil {
			fail("fail to connect to redis server:", redisErr)
		}
		defer redisConn.Close()
		debug("connected to redis:", c.String("redis"), "database number:", c.Int("number"))

		listener, err := net.Listen("tcp", c.String("bind"))
		if err != nil {
			fail("fail to listen:", err)
		}
		defer listener.Close()
		debug("cache-got-hit is listening:", c.String("bind"))

		for {
			cghConn, err := listener.Accept()
			if err != nil {
				debug(err)
				continue
			}
			go process(cghConn)
		}
	}
	app.Run(os.Args)
}
