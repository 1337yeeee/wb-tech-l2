package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

func main() {
	timeout := flag.Int("timeout", 10, "connection timeout in seconds")
	flag.Parse()
	if flag.NArg() < 2 {
		log.Fatal("Usage: telnet <host> <port> [--timeout=N]")
	}

	host := flag.Arg(0)
	port := flag.Arg(1)

	// установка соединения
	conn, err := net.DialTimeout("tcp", host+":"+port, time.Duration(*timeout)*time.Second)
	if err != nil {
		log.Fatalf("Failed to connect to %s:%s: %v", host, port, err)
	}
	defer conn.Close()

	fmt.Printf("Connected to %s:%s\n", host, port)
	fmt.Println("Ctrl+D to exit")

	var wg sync.WaitGroup
	wg.Add(2)

	// stdin → conn
	go func() {
		defer wg.Done()
		if _, err := io.Copy(conn, os.Stdin); err != nil && err != io.EOF {
			log.Println("Error writing to server:", err)
		}
		conn.(*net.TCPConn).CloseWrite() // закрыть запись, чтобы сигнализировать серверу
	}()

	// conn → stdout
	go func() {
		defer wg.Done()
		if _, err := io.Copy(os.Stdout, conn); err != nil && err != io.EOF {
			log.Println("Error reading from server:", err)
		}
	}()

	wg.Wait()
	fmt.Println("\nConnection closed.")
}
