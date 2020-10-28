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
	var PORT string;
	if len(arguments) != 1 {
		if len(arguments) != 2{
			fmt.Println("Please provide a runport and host:port string");
			return;
		}
		PORT = arguments[1]
	} else {
		readerTmp := bufio.NewReader(os.Stdin);
		fmt.Print("enter run Port(<1234>) >> ");
		PORT, _ := readerTmp.ReadString('\n');
		PORT = strings.Trim(PORT, " \n");
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

	// fmt.Printf("The UDP server is %s\n", c.RemoteAddr().String())
	defer c.Close()

	for {
		buffer := make([]byte, 1024)
		n, sender, err := c.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println(err)
			return
		}
		var tt protocol.Test; 
		protocol.DecodeFromBytes(buffer[0:n], &tt, true);
		
		fmt.Print("Received from ", sender,  " >> ", string(tt.Data), "\n");


		
		sendData := protocol.Test{ A: 0.0, B: 0, Data: []byte("ping " + fmt.Sprint("you: ", sender, "\n"))};
		data := protocol.EncodeToBytes(sendData, true);
		for i := 0; i < 10; i++ {
			_, err = c.WriteToUDP(data,sender);
		}
	
	}
	
}
