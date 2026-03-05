package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	ch := make(chan string)
	currentLine := ""
	go func() {
		for {
			buffer := make([]byte, 8)
			n, err := f.Read(buffer)
			if err != nil {
				if errors.Is(err, io.EOF) {
					if currentLine != "" {
						ch <- currentLine
					}
					close(ch)
					break
				}
			}
			currentLine += string(buffer[:n])
			parts := strings.Split(currentLine, "\n")
			for part := range len(parts) - 1 {
				ch <- parts[part]
			}
			currentLine = parts[len(parts)-1]
		}
		f.Close()
	}()
	return ch
}

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for {
		con, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Printf("connection accepted\n")
		ch := getLinesChannel(con)
		for line := range ch {
			fmt.Printf("%s\n", line)
		}
		fmt.Printf("connection closed\n")
	}
}
