package protocol

import (
	// "fmt"
	"bytes"
	"encoding/gob"
	"log"
	"compress/gzip"
	"io/ioutil"
)

//Test test type
type Test struct {
	A float32
	B uint16
	Data []byte
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