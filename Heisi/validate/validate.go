package main

import (
	"bufio"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/bits"
	"os"
	"strconv"
)

type item [10]byte

const (
	template2021    = "http://hs.heisiwu.com/wp-content/uploads/%4d/%02d/%4d%02d16%06d-611a3%8s.jpg"
	templategeneral = "http://hs.heisiwu.com/wp-content/uploads/%4d/%02d/%015x"
)

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
	ouf, err := os.Open(os.Args[2])
	if err != nil {
		panic(err)
	}
	defer ouf.Close()
	s := bufio.NewScanner(inf)
	i := 0
	for s.Scan() {
		i++
		t := s.Text()
		var it item
		ouf.Read(it[:])
		year, month := int((it[0]>>4)&0x0f), int(it[0]&0x0f)
		year += 2021
		if year == 2021 {
			num := binary.BigEndian.Uint32(it[1:5])
			dstr := hex.EncodeToString(it[5:9])
			trestore := fmt.Sprintf(template2021, year, month, year, month, num, dstr)
			if trestore != t {
				panic("line " + strconv.Itoa(i) + ": mismatched content " + trestore)
			}
			continue
		}
		d := binary.BigEndian.Uint64(it[1:9])
		isscaled := it[9]&0x80 > 0
		num := int(it[9] & 0x7f)
		trestore := fmt.Sprintf(templategeneral, year, month, d&0x0fffffff_ffffffff)
		if num > 0 {
			trestore += fmt.Sprintf("-%d", num)
		}
		if isscaled {
			trestore += "-scaled"
		}
		d = bits.RotateLeft64(d, 4) & 0x0f
		switch d {
		case 0:
			trestore += ".jpg"
		case 1:
			trestore += ".png"
		case 2:
			trestore += ".webp"
		default:
			panic("line " + strconv.Itoa(i) + ": invalid ext")
		}
		if trestore != t {
			panic("line " + strconv.Itoa(i) + ": mismatched content " + trestore)
		}
	}
}
