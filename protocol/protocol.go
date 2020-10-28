package protocol

import (
	// "fmt"
	"bytes"
	"encoding/gob"
	"log"
	"compress/gzip"
	"io/ioutil"
	"net"
	"time"
	"sync"
	"fmt"
)

import "hash/crc32" 

//Test test type
type Test struct {
	A float32
	B uint16
	Data []byte
}

//PackageType for packages
type PackageType byte

/*
ACK acknowledge

*/
const (
	EnumPackageACK PackageType = iota
	EnumPackageData
)

//Package Package type, number, and payload
type Package struct {
	Type PackageType
	Number uint64 
	Payload []byte
	Crc32 uint32 // probably UDP 16bit checksum is not enough
}

//CalcChecksum calculates and write checksum
func (pkg *Package) CalcChecksum(){
	var allBytes []byte;
	allBytes = append(allBytes, byte(pkg.Type));
	allBytes = append(allBytes, Uint64ToByteSlice(pkg.Number)...);
	allBytes = append(allBytes, pkg.Payload...);

	pkg.Crc32 = crc32.ChecksumIEEE(allBytes);
}

//CheckChecksum calculates and check checksum
func (pkg *Package) CheckChecksum() bool{
	var allBytes []byte;
	allBytes = append(allBytes, byte(pkg.Type));
	allBytes = append(allBytes, Uint64ToByteSlice(pkg.Number)...);
	allBytes = append(allBytes, pkg.Payload...);

	if pkg.Crc32 == crc32.ChecksumIEEE(allBytes) {
		return true;
	}
	return false;
}

//Peer struct for peers
type Peer struct {
	Addr 					*net.UDPAddr
	
	PackageNumber 			uint64

	AckQueue				chan uint64

	LastSent 				*Package
	LastSentMutex			sync.Mutex

	ReceivedDataChannel 	chan []byte
	SendPayloadChannel 		chan []byte
	ReceiverPayloadChannel 	chan []byte
	UDPDataChannel 			chan []byte

	SyncChan				chan *Package
}

//InitializePeer initialize peer with Address of peer and UDP channel 
func (p *Peer) InitializePeer(Addr *net.UDPAddr, UDPDataChannel chan []byte){
	p.Addr = Addr

	p.PackageNumber = 0;

	p.AckQueue	= make(chan uint64, 100);

	p.LastSent = nil;

	p.ReceivedDataChannel = make(chan []byte, 100);
	p.SendPayloadChannel = make(chan []byte, 100);
	p.ReceiverPayloadChannel = make(chan []byte, 100);
	p.UDPDataChannel = UDPDataChannel;

	p.SyncChan = make(chan *Package, 1);



	go p.Controller();
	go p.senderF();
	go p.receiverF();

	fmt.Print("initilization, done\n");
}

var counter = 0

//Controller control the acks and new packages
func (p *Peer) Controller(){
	outerLoop:
	for {
		// p.LastSentMutex.Lock();
		timeouter := time.After(100 * time.Millisecond)

		if p.LastSent != nil {
			innerLoop:
			for {
				select {
				case <- timeouter:
					//Timeout
					fmt.Println("timeout, resending ", counter)
					counter++
					p.SyncChan <- p.LastSent
					continue outerLoop
				case ackNumber := <- p.AckQueue:
					counter = 0
					if ackNumber == p.LastSent.Number {
						p.LastSent = nil
						break innerLoop	
					}
				}
			}
		}

		newPayload := <-p.SendPayloadChannel
		newPackage := Package{Type: EnumPackageData, Number: p.PackageNumber, Payload: newPayload};
		p.PackageNumber++;
		newPackage.CalcChecksum();

		p.LastSent = &newPackage;
		p.SyncChan <- p.LastSent
			
		// p.LastSentMutex.Unlock();
	}

} 

func (p *Peer) senderF() {
	for pkg := range p.SyncChan {
		p.UDPDataChannel <- EncodeToBytes(*pkg, true);
		// go p.timeOut(pkg.Number, 100);
	}
}

func (p *Peer) receiverF() {
	for data := range p.ReceivedDataChannel {
		var pkg Package;
		DecodeFromBytes(data, &pkg, true);
		
		if pkg.CheckChecksum() != true {
			fmt.Println("Wrong checksum")
			continue;
		}

		// fmt.Print("pkg: ", pkg);

		if pkg.Type == EnumPackageACK {
				p.AckQueue <- pkg.Number 		
		} else {

			ack := Package{Type: EnumPackageACK, Number: pkg.Number, Payload: []byte{}};
			ack.CalcChecksum()
			p.UDPDataChannel <- EncodeToBytes(ack, true);

			switch pkg.Type {
				case EnumPackageACK:
				case EnumPackageData:
					p.ReceiverPayloadChannel <- pkg.Payload
				default: 
			}
		}
	}
}

//Uint64ToByteSlice convert uint64 to []byte
func Uint64ToByteSlice(v uint64) (b []byte) {
	b = []byte{
		byte(0xff & v),
		byte(0xff & (v >> 8)),
		byte(0xff & (v >> 16)),
		byte(0xff & (v >> 24)),
		byte(0xff & (v >> 32)),
		byte(0xff & (v >> 40)),
		byte(0xff & (v >> 48)),
		byte(0xff & (v >> 56))}
	return;
}


//EncodeToBytes encodes a struct to []byte
func EncodeToBytes(p interface{}, compress bool) []byte {

	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(p)
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Println("uncompressed size (bytes): ", len(buf.Bytes()))

	if compress {
		return Compress(buf.Bytes());
	}

	return buf.Bytes()
}

/*
Compress get a []byte and return compressed []byte
Compressed []byte might be larger
*/
func Compress(s []byte) []byte {

	zipbuf := bytes.Buffer{}
	zipped := gzip.NewWriter(&zipbuf)
	zipped.Write(s)
	zipped.Close()
	// fmt.Println("compressed size (bytes): ", len(zipbuf.Bytes()))
	return zipbuf.Bytes()
}


//Decompress get a []byte, and return decompressed []byte
func Decompress(s []byte) []byte {
	rdr, _ := gzip.NewReader(bytes.NewReader(s))
	data, err := ioutil.ReadAll(rdr)
	if err != nil {
		log.Fatal(err)
	}
	rdr.Close()
	// fmt.Println("uncompressed size (bytes): ", len(data))
	return data
}

//DecodeFromBytes gets bytes and a pointer for a struct  to []byte
func DecodeFromBytes(s []byte, p interface{}, decompress bool) {
	if decompress {
		s = Decompress(s);
	}
	err := gob.NewDecoder(bytes.NewReader(s)).Decode(p);
	if err != nil {
		log.Fatal(err)
	}
}

//thanks for https://gist.github.com/SteveBate/042960baa7a4795c3565