package client

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

type Config struct {
	Host    string
	Port    string
	Timeout time.Duration
}

func Start(conf Config) {

	var timeout = conf.Timeout

	dialer := &net.Dialer{
		Timeout: timeout,
	}
	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	conn, err := dialer.DialContext(ctx, "tcp", conf.Host+":"+conf.Port)

	if err != nil {
		log.Fatalf("Cannot connect: %v", err)
	}

	wg.Add(1)
	go func() {
		readSocket(ctx, cancel, conn)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		writeSocket(ctx, conn)
		wg.Done()
	}()

	wg.Wait()
	conn.Close()
}

func readSocket(ctx context.Context, cancel context.CancelFunc, conn net.Conn) {
	scanner := bufio.NewScanner(conn)

	for {
		select {
		case <-ctx.Done():
			break
		default:
			if !scanner.Scan() {
				cancel()
				conn.Close()
				log.Fatal("connection refused")
				break
			}
			text := scanner.Text()
			log.Printf("read from server %s", text)
		}
	}

	log.Printf("finished reading")
}

func writeSocket(ctx context.Context, conn net.Conn) {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		select {
		case <-ctx.Done():
			break
		default:
			if !scanner.Scan() {
				conn.Close()
				log.Fatal("connection refused")
				break
			}
			text := scanner.Text()
			log.Printf("write to server %s", text)
			conn.Write([]byte(fmt.Sprintf("%s\n", text)))
		}
	}
	log.Printf("finished writing")
}
