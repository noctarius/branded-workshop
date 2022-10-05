package main

import (
	"bufio"
	"encoding/binary"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
)

const format = "2006-01-02 15:04:05.999999999-07"

func main() {
	i, err := os.Open("./export.csv")
	if err != nil {
		panic(err)
	}
	defer i.Close()

	o, err := os.OpenFile("./data.bin", os.O_CREATE|os.O_TRUNC|os.O_RDWR, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer o.Close()

	r := bufio.NewReader(i)
	_, _, err = r.ReadLine()
	if err != nil {
		panic(err)
	}

	/*c, err := gzip.NewWriterLevel(o, gzip.BestCompression)
	if err != nil {
		panic(err)
	}
	defer c.Close()*/

	for {
		l, p, err := r.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}

		if p {
			panic("ARGH!")
		}

		line := string(l)
		tokens := strings.Split(line, ",")

		timestamp, err := time.Parse(format, tokens[0])
		if err != nil {
			panic(err)
		}
		millis := timestamp.UnixMilli()

		typeId, err := strconv.ParseInt(tokens[1], 10, 32)
		if err != nil {
			panic(err)
		}

		value, err := strconv.ParseFloat(tokens[2], 64)
		if err != nil {
			panic(err)
		}

		//fmt.Printf("time=%d, type=%d, value=%0.2f\n", millis, typeId, value)

		data := make([]byte, 17, 17)
		bits := math.Float64bits(value)
		t := math.Float64frombits(bits)
		if t != value {
			panic("WUT")
		}

		binary.LittleEndian.PutUint64(data[0:], uint64(millis))
		data[8] = uint8(typeId)
		binary.LittleEndian.PutUint64(data[9:], bits)

		if _, err := o.Write(data); err != nil {
			panic(err)
		}
	}
}
