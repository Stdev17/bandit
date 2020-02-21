package parse

import (
	."fmt"
	"time"
	//"bufio"
	"io"
	"io/ioutil"
	"os"
	"errors"
	"encoding/binary"
	"bytes"
	"compress/zlib"
)

// Each chunk starts with a uint32 length (big endian), then 4 byte name,
// then data and finally the CRC32 of the chunk data.
type Chunk struct {
    Length int    // chunk data length
    CType  string // chunk type
    Data   []byte // chunk data
    Crc32  []byte // CRC32 of chunk data
}

type PNG struct {
    Width             int
    Height            int
    BitDepth          int
    ColorType         int
    CompressionMethod int
    FilterMethod      int
    InterlaceMethod   int
    chunks            []*Chunk // Not exported == won't appear in JSON string.
    NumberOfChunks    int
}

// uInt32ToInt converts a 4 byte big-endian buffer to int.
func uInt32ToInt(buf []byte) (int, error) {
    if len(buf) == 0 || len(buf) > 4 {
        return 0, errors.New("invalid buffer")
    }
    return int(binary.BigEndian.Uint32(buf)), nil
}

// Populate will read bytes from the reader and populate a chunk.
func (c *Chunk) Populate(buf* []byte, ptr int) (int, error) {

    var err error
    c.Length, err = uInt32ToInt((*buf)[ptr:ptr+4])
    if err != nil {
        return ptr, errors.New("cannot convert length to int")
    }

	c.CType = string((*buf)[ptr+4:ptr+8])

	c.Data = (*buf)[ptr+8:ptr+8+c.Length]
	
	ptr += (12+c.Length)

	c.Crc32 = (*buf)[ptr-4:ptr]

    return ptr, nil
}

// Parse IHDR chunk.
// https://golang.org/src/image/png/reader.go?#L142 is your friend.
func (png *PNG) parseIHDR(iHDR *Chunk) error {
	iHDRlength := 13
    if iHDR.Length != iHDRlength {
        errString := Sprintf("invalid IHDR length: got %d - expected %d",
            iHDR.Length, iHDRlength)
        return errors.New(errString)
    }

    tmp := iHDR.Data
    var err error

    png.Width, err = uInt32ToInt(tmp[0:4])
    if err != nil || png.Width <= 0 {
        errString := Sprintf("invalid width in iHDR - got %x", tmp[0:4])
        return errors.New(errString)
    }

    png.Height, err = uInt32ToInt(tmp[4:8])
    if err != nil || png.Height <= 0 {
        errString := Sprintf("invalid height in iHDR - got %x", tmp[4:8])
        return errors.New(errString)
    }

    png.BitDepth = int(tmp[8])
    png.ColorType = int(tmp[9])

    // Only compression method 0 is supported
    if int(tmp[10]) != 0 {
        errString := Sprintf("invalid compression method - expected 0 - got %x",
            tmp[10])
        return errors.New(errString)
    }
    png.CompressionMethod = int(tmp[10])

    // Only filter method 0 is supported
    if int(tmp[11]) != 0 {
        errString := Sprintf("invalid filter method - expected 0 - got %x",
            tmp[11])
        return errors.New(errString)
    }
    png.FilterMethod = int(tmp[11])

    // Only interlace methods 0 and 1 are supported
    if int(tmp[12]) != 0 {
        errString := Sprintf("invalid interlace method - expected 0 or 1 - got %x",
            tmp[12])
        return errors.New(errString)
    }
    png.InterlaceMethod = int(tmp[12])

    return nil
}

func (png *PNG) Add(c *Chunk) {
	png.chunks = append(png.chunks, c)
}

func ZlibDecom(idat* []byte, dest *bytes.Buffer) {
	b := bytes.NewReader((*idat))
	r, err := zlib.NewReader(b)
	if err != nil {
		panic(err)
	}
	io.Copy(dest, r)
}

func ParseMain() {
	
	f, _ := os.Open("/home/leta/misc/bandit/resources/kokkoro2.png")

	//r := bufio.NewReaderSize(f, 1<<25)
	buf, _ := ioutil.ReadAll(f)

	start := time.Now()

	ptr := 8

	var png PNG
	var ihdr Chunk
	ptr, _ = ihdr.Populate(&buf, ptr);
	png.parseIHDR(&ihdr);
	png.Add(&ihdr);

    for {
        var c Chunk
		ptr, _ = c.Populate(&buf, ptr);
		png.Add(&c);
        if c.CType == "IEND" {
            break;
        }
	}

	du := time.Since(start)
	Println(du)

	r := new(bytes.Buffer)
	ZlibDecom(&(png.chunks[len(png.chunks)-2].Data), r);

	du = time.Since(start)
	Println(du)
}