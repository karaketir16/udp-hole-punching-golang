package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"my.test.com/protocol"
)



func main() {
	arguments := os.Args
	var PORT, SERVER string;
	if len(arguments) != 1 {
		if len(arguments) != 3{
			fmt.Println("Please provide a runport and host:port string");
			return;
		}
		PORT = arguments[1]
		SERVER = arguments[2]
	} else {
		readerTmp := bufio.NewReader(os.Stdin);
		fmt.Print("enter run Port(<1234>) >> ");
		PORT, _ := readerTmp.ReadString('\n');
		PORT = strings.Trim(PORT, " \n");
	
	
	
		readerTmp = bufio.NewReader(os.Stdin)
		fmt.Print("enter server(<34.23.112.11:1234>) >> ");
		SERVER, _ := readerTmp.ReadString('\n')
		SERVER = strings.Trim(SERVER, " \n")
	}


	s, err := net.ResolveUDPAddr("udp4", ":" + PORT)
	if err != nil {
		fmt.Println(err)
		return
	}

	c, err := net.ListenUDP("udp4", s)
	if err != nil {
		fmt.Println(err)
		return
	}

	serverAddr, err := net.ResolveUDPAddr("udp4", SERVER);
	if err != nil {
		fmt.Println(err)
		return
	}

	go func(){
		sendData := protocol.Test{ A: 0.0, B: 0, Data: []byte("ping")};
		data := protocol.EncodeToBytes(sendData, true);
		for i := 0; i < 10; i++ {
			_, err = c.WriteToUDP(data,serverAddr);
		}
	}()


	buffer := make([]byte, 1024)
	n, sender, err := c.ReadFromUDP(buffer)
	if err != nil {
		fmt.Println(err)
		return
	}
	var tt protocol.Test; 
	protocol.DecodeFromBytes(buffer[0:n], &tt, true);
	fmt.Print("Received from ", sender,  " >> ", string(tt.Data));


	readerTmp := bufio.NewReader(os.Stdin)
	fmt.Print("enter receiver(<34.23.112.11:1234>) >> ");
	textTmp, _ := readerTmp.ReadString('\n')

	addr, err := net.ResolveUDPAddr("udp4", strings.Trim(textTmp, " \n"));

	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("addr", addr);
	fmt.Println("serverAddr", serverAddr);
	fmt.Println("SERVER", SERVER);

	// fmt.Printf("The UDP server is %s\n", c.RemoteAddr().String())
	defer c.Close()

	go func() {
		for {
			buffer := make([]byte, 1024)
			n, sender, err := c.ReadFromUDP(buffer)
			if err != nil {
				fmt.Println(err)
				return
			}
			var tt protocol.Test; 
			protocol.DecodeFromBytes(buffer[0:n], &tt, true);
			
			fmt.Print("Received from ", sender,  " >> ", string(tt.Data));
		}
	}()

	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print(">> ")
		text, _ := reader.ReadString('\n')

		sendData := protocol.Test{ A: 0.0, B: 0, Data: []byte(text)};
		data := protocol.EncodeToBytes(sendData, true);
		_, err = c.WriteToUDP(data,addr)
		if strings.TrimSpace(string(data)) == "STOP" {
			fmt.Println("Exiting UDP client!")
			return
		}

		if err != nil {
			fmt.Println(err)
			return
		}
	}
}
