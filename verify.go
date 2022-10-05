package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
)

const format = "2006-01-02 15:04:05.999999999-07"

func main() {
	i1, err := os.Open("./export.csv")
	if err != nil {
		panic(err)
	}
	defer i1.Close()

	i2, err := os.Open("./data.bin")
	if err != nil {
		panic(err)
	}
	defer i2.Close()

	r1 := bufio.NewReader(i1)
	_, _, err = r1.ReadLine()
	if err != nil {
		panic(err)
	}

	for {
		l, p, err := r1.ReadLine()
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

		data := make([]byte, 17, 17)
		bits := math.Float64bits(value)

		binary.LittleEndian.PutUint64(data[0:], uint64(millis))
		data[8] = uint8(typeId)
		binary.LittleEndian.PutUint64(data[9:], bits)

		data2 := make([]byte, 17, 17)
		if _, err := i2.Read(data2); err != nil {
			panic(err)
		}

		v := binary.LittleEndian.Uint64(data2[0:])
		millis2 := int64(v)

		typeId2 := int64(data2[8])

		v = binary.LittleEndian.Uint64(data2[9:])
		value2 := math.Float64frombits(v)

		if !bytes.Equal(data, data2) {
			fmt.Printf("data=%X, data2=%X\n", data, data2)
			fmt.Printf("time=%d, type=%d, value=%0.2f => time=%d, type=%d, value=%0.2f\n", millis, typeId, value, millis2, typeId2, value2)
			panic("bam")
		}

		if (millis != millis2) || (typeId != typeId) || (value != value2) {
			fmt.Printf("time=%d, type=%d, value=%0.2f => time=%d, type=%d, value=%0.2f\n", millis, typeId, value, millis2, typeId2, value2)
		}
	}
}
