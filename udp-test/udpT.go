package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"my.test.com/protocol"
	"time"
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


	s, err := net.ResolveUDPAddr("udp4", "127.0.0.1:" + PORT)
	if err != nil {
		fmt.Println(err)
		return
	}

	c, err := net.ListenUDP("udp4", s)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer c.Close()

	serverAddr, err := net.ResolveUDPAddr("udp4", SERVER);
	if err != nil {
		fmt.Println(err)
		return
	}



	udpChannel := make(chan []byte) 

	peerMap := make(map[string]*protocol.Peer)

	peer1 := protocol.Peer{}
	peer1.InitializePeer(serverAddr, udpChannel)
	peerMap[peer1.Addr.String()] = &peer1
	
	go func(){
		for data := range udpChannel {
			c.WriteToUDP(data, peer1.Addr);
		}
	}()

	go func(){
		for data := range peer1.ReceiverPayloadChannel {
			fmt.Println("Received\n    as byte: ", data, "\n    as string: ", string(data))
		}
	}()

	pinger := time.NewTicker(1000 * time.Millisecond)

    go func() {
        for range pinger.C{
            peer1.SendPayloadChannel <- []byte("ping")
        }
    }()


	func() {
		for {
			buffer := make([]byte, 1024)
			n, sender, err := c.ReadFromUDP(buffer)
			if err != nil {
				fmt.Println(err)
				return
			}
			peerMap[sender.String()].ReceivedDataChannel <- buffer[0:n];
		}
	}()

	// fmt.Printf("The UDP server is %s\n", c.RemoteAddr().String())
	
}
