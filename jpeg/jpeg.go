// Package jpeg implements decoding of jpeg file
package jpeg

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
)

/*
https://github.com/Matthias-Wandel/jhead/blob/master/jhead.h
#define M_SOF0  0xC0          // Start Of Frame N
#define M_SOF1  0xC1          // N indicates which compression process
#define M_SOF2  0xC2          // Only SOF0-SOF2 are now in common use
#define M_SOF3  0xC3
#define M_SOF5  0xC5          // NB: codes C4 and CC are NOT SOF markers
#define M_SOF6  0xC6
#define M_SOF7  0xC7
#define M_SOF9  0xC9
#define M_SOF10 0xCA
#define M_SOF11 0xCB
#define M_SOF13 0xCD
#define M_SOF14 0xCE
#define M_SOF15 0xCF
#define M_SOI   0xD8          // Start Of Image (beginning of datastream)
#define M_EOI   0xD9          // End Of Image (end of datastream)
#define M_SOS   0xDA          // Start Of Scan (begins compressed data)
#define M_JFIF  0xE0          // Jfif marker
#define M_EXIF  0xE1          // Exif marker.  Also used for XMP data!
#define M_XMP   0x10E1        // Not a real tag (same value in file as Exif!)
#define M_COM   0xFE          // COMment
#define M_DQT   0xDB          // Define Quantization Table
#define M_DHT   0xC4          // Define Huffmann Table
#define M_DRI   0xDD
#define M_IPTC  0xED

SOI	0xFF, 0xD8			none			Start of Image
S0F0	0xFF, 0xC0		variable size	Start of Frame
S0F2	0xFF, 0xC2		variable size	Start fo Frame
DHT	0xFF, 0xC4			variable size	Define Huffman Tables
DQT	0xFF, 0xDB			variable size	Define Quantization Table(s)
DRI	0xFF, 0xDD			4 bytes	Define Restart Interval
SOS	0xFF, 0xDA			variable size	Start Of Scan
APPn	0xFF, 0xE//n//	variable size	Application specific
COM	0xFF, 0xFE			variable size	Comment
RSTn	0xFF, 0xD//n//(//n//#0..7)	none	Restart
EOI	0xFF, 0xD9			none End Of Image
*/
const (
	//after 0xFF
	ff_SOI  = 0xD8 // start of image
	ff_SOF0 = 0xC0 // start of frame
	ff_SOF2 = 0xC2 // start of frame
	ff_DHT  = 0xC4 // define Huffman tables
	ff_DQT  = 0xDB // define Quantization table(s)
	ff_DRI  = 0xDD // 4 bytes define restart interval
	ff_EOI  = 0xD9 // end of image
	ff_SOS  = 0xDA // start of scan
	ff_RST0 = 0xD0 // restart 0
	ff_RST1 = 0xD1 // restart 1
	ff_RST2 = 0xD2 // restart 2
	ff_RST3 = 0xD3 // restart 3
	ff_RST4 = 0xD4 // restart 4
	ff_RST5 = 0xD5 // restart 5
	ff_RST6 = 0xD6 // restart 6
	ff_RST7 = 0xD7 // restart 7
	ff_APP0 = 0xE0 // application #0
	ff_APP1 = 0xE1 // EXIF, next two byte are SIZE
	ff_APP2 = 0xE2 // application #2
	ff_APP4 = 0xE4 // application #4
	ff_APPD = 0xED // application #D
	ff_COMM = 0xFE // comment,  next two byte are SIZE

)

type JpegReader struct {
	Filename    string
	br          *bufio.Reader
	ImageHeight int
	ImageWidth  int
	Comment     string
	verbose     bool
	maxCounter  int // just to develop and debug
	charCounter int
}

func (j *JpegReader) Open(fname string, verbose bool) error {
	j.maxCounter = 0x6dbe + 0x0a00 + 320000000 // just to develop and debug
	file, err := os.Open(fname)
	j.Filename = fname
	if err != nil {
		return err
	}
	j.br = bufio.NewReader(file)
	j.verbose = verbose
	return nil
}

func (j *JpegReader) NextByte() byte {
	b, err := j.br.ReadByte()
	if err != nil {
		fmt.Printf("error reading file %s\n", j.Filename)
		panic(err)
	}
	//if j.charCounter > j.maxCounter {
	//	pa nic("char limit over")
	//}
	j.charCounter++
	return b
}

func (j *JpegReader) SegmentLength() int {
	dataLenBytes := make([]byte, 2)
	dataLenBytes[0] = j.NextByte()
	dataLenBytes[1] = j.NextByte()
	return int(binary.BigEndian.Uint16(dataLenBytes))
}
func (j *JpegReader) Bytes0(length int) ([]byte, error) {
	var data []byte
	data = make([]byte, length)
	for i := 0; i < length; i++ {
		data[i] = j.NextByte()
	}
	return data, nil
}
func (j *JpegReader) Bytes(length int) ([]byte, error) {
	var data []byte
	nread := 0
	for nread < length {
		s := make([]byte, length-nread)
		n, err := j.br.Read(s)
		nread += n
		if err != nil { //&& nread < length {
			j.charCounter += nread
			return nil, err
		}
		data = append(data, s[:n]...)
	}
	j.charCounter += nread
	return data, nil
}
func (j *JpegReader) PrintMark(markName string, m byte) {
	if j.verbose {
		fmt.Printf("\n%06x %s %02x ", j.charCounter, markName, m)
	}
}

func (j *JpegReader) PrintMarkAndGetData(markName string, m byte) (int, []byte) {
	j.PrintMark(markName, m)
	length := j.SegmentLength()
	data, err := j.Bytes(length - 2)
	if err != nil {
		fmt.Printf("error at file %s: %s\n", j.Filename, err.Error())
	}
	return length, data
}

func (j *JpegReader) PrintMarkLength(markName string, m byte) {
	length, data := j.PrintMarkAndGetData(markName, m)
	if j.verbose {
		fmt.Printf(" %04x -2=%d ", length, len(data))
	}
}

func (j *JpegReader) PrintMarkComment(markName string, m byte) {
	length, data := j.PrintMarkAndGetData(markName, m)
	j.Comment = string(data)
	if j.verbose {
		fmt.Printf(" %04x -2=%d %s", length, len(data), j.Comment)
	}
}
func (j *JpegReader) PrintMarkStartOfFrame0(markName string, m byte) {
	length, data := j.PrintMarkAndGetData(markName, m)
	j.ImageHeight = int(binary.BigEndian.Uint16(data[1:3]))
	j.ImageWidth = int(binary.BigEndian.Uint16(data[3:5]))
	if j.verbose {
		fmt.Printf(" %04x -2=%d %s", length, len(data), j.Comment)
	}
}

func (j *JpegReader) ScanToFFmark() byte {
	c := j.NextByte()
	if j.verbose {
		fmt.Printf("\n%06x SSS %02x = char after ff,  Scan now\n", j.charCounter, c)
	}
	for {
		for c != 0xff {
			c = j.NextByte()
		}
		switch m := j.NextByte(); m {
		case 0xFF:
			j.PrintMark("FFF", m)
			c = j.NextByte()
			if c == 0x00 {
				c = j.NextByte()
			}
		case ff_RST0, ff_RST1, ff_RST2, ff_RST3, ff_RST4, ff_RST5, ff_RST6, ff_RST7: //restart
			//j.PrintMark("RST", m)
			c = j.NextByte()
		case 0x00:
			c = j.NextByte()
		default:
			return m
		}
	}
}

func (j *JpegReader) Decode() error {
	cff := j.NextByte()
	if cff != 0xFF {
		return fmt.Errorf("file %s is not JPEG, does not start by 0xFF, but %02x", j.Filename, cff)
	}
	m := j.NextByte()
	if m != ff_SOI {
		return fmt.Errorf("not JPEG, does not start by 0xFF%02x, but 0xFF%02x", ff_SOI, m)
	}
	cff = j.NextByte()
	scan := false
	for cff == 0xff {
		if scan {
			m = j.ScanToFFmark()
			scan = false
		} else {
			m = j.NextByte()
		}
		switch m {
		case ff_SOF0, ff_SOF2: //start of frame
			j.PrintMarkStartOfFrame0("SOF", m)
		case ff_DHT: // define Huffman tables
			j.PrintMarkLength("DHT", m)
		case ff_DQT: // define Quantization table(s)
			j.PrintMarkLength("DQT", m)
		case ff_DRI: // 4 bytes define restart interval
			j.PrintMark("DRI", m)
		case ff_APP0: // application #0
			j.PrintMark("AP0", m)
		case ff_APP1: // EXIF, next two byte are SIZE
			j.PrintMarkLength("AP1", m)
		case ff_APP2: // application #2
			j.PrintMark("AP2", m)
		case ff_APP4: // application #4
			j.PrintMark("AP4", m)
		case ff_APPD: // application #D
			j.PrintMark("APD", m)
		case ff_COMM: // comment,  next two byte are SIZE
			j.PrintMarkComment("COM", m)
		case ff_SOS: // start of scan
			j.PrintMarkLength("SOS", m)
			scan = true
			return nil // let's skip scan
		case ff_EOI: // end of image
			j.PrintMark("EOI", m)
			//return
		default:
			store := j.verbose
			j.verbose = true
			j.PrintMark("???", m)
			j.verbose = store
		}
		if m == ff_EOI {
			break
		}
		if !scan {
			cff = j.NextByte()
		}
	}
	if j.verbose {
		fmt.Printf("\n%06x KON %02x KON-EC SMYCKY scan=%t\n", j.charCounter, cff, scan)
	}
	return nil
}

/* 2023-0412
Arcpics - created:    4f   0s /media/josef/JosefsPassport1TB/Arc-Pics/Slides/arcpics.json
Arcpics - created:    6f   0s /media/josef/JosefsPassport1TB/Arc-Pics/Sounds/arcpics.json
Elapsed time 16m7.591727739s
panic: EOF

goroutine 1 [running]:
github.com/chlachula/arcpics/jpeg.(*JpegReader).NextByte(0xc00008d3f0)
	/home/josef/go/josef/arcpics/jpeg/jpeg.go:103 +0x4e
github.com/chlachula/arcpics/jpeg.(*JpegReader).Decode(0xc00008d3f0)
	/home/josef/go/josef/arcpics/jpeg/jpeg.go:209 +0x33
github.com/chlachula/arcpics.getJpegComment({0xc0010db220?, 0x2?})
	/home/josef/go/josef/arcpics/file.go:260 +0x54
github.com/chlachula/arcpics.makeJdir({0xc0001b4c80, 0x38})
	/home/josef/go/josef/arcpics/file.go:293 +0x679
github.com/chlachula/arcpics.ArcpicsFilesUpdate.func1({0xc0001b4c80, 0x38}, {0x533dc0?, 0xc0000bb000?}, {0x0?, 0x0?})
	/home/josef/go/josef/arcpics/file.go:190 +0x18a
path/filepath.walkDir({0xc0001b4c80, 0x38}, {0x533dc0, 0xc0000bb000}, 0xc00008dda0)
	/usr/local/go/src/path/filepath/path.go:445 +0x5c
path/filepath.walkDir({0x7ffcb52cb3f7, 0x27}, {0x533df8, 0xc0000a8020}, 0xc00008dda0)
	/usr/local/go/src/path/filepath/path.go:467 +0x2a7
path/filepath.WalkDir({0x7ffcb52cb3f7, 0x27}, 0xc00004dda0)
	/usr/local/go/src/path/filepath/path.go:535 +0xb0
github.com/chlachula/arcpics.ArcpicsFilesUpdate({0x7ffcb52cb3f7, 0x27})
	/home/josef/go/josef/arcpics/file.go:179 +0xf0
main.update_dirs_or_db(0x0?, 0x1, {0x50d641?, 0xc00002a9b0?})
	/home/josef/go/josef/arcpics/cmd/main.go:57 +0xc5
main.main()
	/home/josef/go/josef/arcpics/cmd/main.go:100 +0x510
exit status 2

*/
