package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	udp, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		log.Fatalf("error: %s\n", err.Error())
	}
	udpCon, err := net.DialUDP("udp", nil, udp)
	if err != nil {
		log.Fatalf("error: %s\n", err.Error())
	}
	defer udpCon.Close()
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println(">")
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("error: %s\n", err.Error())
		}
		udpCon.Write([]byte(line))
	}
}
