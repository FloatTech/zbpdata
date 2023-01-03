package main

import (
	"bufio"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type item [10]byte

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: heisi.txt heisi.bin")
		return
	}
	inf, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer inf.Close()
	ouf, err := os.Create(os.Args[2])
	if err != nil {
		panic(err)
	}
	defer ouf.Close()
	s := bufio.NewScanner(inf)
	i := 0
	for s.Scan() {
		i++
		t := s.Text()
		if !strings.HasPrefix(t, "http://hs.heisiwu.com/wp-content/uploads/") {
			panic("line " + strconv.Itoa(i) + ": invalid url prefix")
		}
		t = t[41:]
		if len(t) < 27 {
			panic("line " + strconv.Itoa(i) + ": invalid url suffix")
		}
		year, err := strconv.Atoi(t[:4]) // 4bits
		if err != nil {
			panic("line " + strconv.Itoa(i) + ": " + err.Error())
		}
		if year < 2021 {
			panic("line " + strconv.Itoa(i) + ": invalid year")
		}
		mounth, err := strconv.Atoi(t[5:7]) // 4bits
		if err != nil {
			panic("line " + strconv.Itoa(i) + ": " + err.Error())
		}
		if mounth == 0 || mounth > 12 {
			panic("line " + strconv.Itoa(i) + ": invalid mounth")
		}
		var it item
		it[0] = byte((year-2021)<<4) | byte(mounth&0x0f) // 1byte
		if year == 2021 {
			num, err := strconv.Atoi(t[16 : 16+6]) // 4bytes
			if err != nil {
				panic("line " + strconv.Itoa(i) + ": " + err.Error())
			}
			d, err := hex.DecodeString(t[28 : 28+8]) // 4bytes
			if err != nil {
				panic("line " + strconv.Itoa(i) + ": " + err.Error())
			}
			if len(d) != 4 {
				panic("line " + strconv.Itoa(i) + ": invalid data")
			}
			binary.BigEndian.PutUint32(it[1:], uint32(num))
			copy(it[5:], d)
		} else {
			head := "0"
			if strings.Contains(t[23:], ".png") {
				head = "1"
			} else if strings.Contains(t[23:], ".webp") {
				head = "2"
			}
			d, err := hex.DecodeString(head + t[8:8+15]) // 8bytes
			if err != nil {
				panic("line " + strconv.Itoa(i) + ": " + err.Error())
			}
			if len(d) != 8 {
				panic("line " + strconv.Itoa(i) + ": invalid data")
			}
			copy(it[1:], d)
			if strings.Contains(t[23:], "scaled") {
				it[9] = 0x80
			}
			if t[23] == '-' && t[24] != 's' {
				switch {
				case t[25] == '-' || t[25] == '.':
					it[9] |= (t[24] - '0') & 0x7f
				case t[26] == '-' || t[26] == '.':
					num, err := strconv.Atoi(t[24:26]) // 1byte
					if err != nil {
						panic("line " + strconv.Itoa(i) + ": " + err.Error())
					}
					it[9] |= byte(num) & 0x7f
				case t[27] == '-' || t[27] == '.':
					num, err := strconv.Atoi(t[24:27]) // 1byte
					if err != nil {
						panic("line " + strconv.Itoa(i) + ": " + err.Error())
					}
					it[9] |= byte(num) & 0x7f
				default:
					panic("line " + strconv.Itoa(i) + ": invalid num")
				}
			}
		}
		ouf.Write(it[:])
	}
}
