//go:build ignore
// +build ignore

package main

import (
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"unicode"
	"unsafe"

	"github.com/FloatTech/floatbox/process"
	"github.com/fumiama/go-registry"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"
)

func main() {
	c := flag.Bool("c", false, "check repo only")
	flag.Parse()
	es, err := fs.ReadDir(os.DirFS("./"), ".")
	if err != nil {
		panic(err)
	}
	for _, es := range es {
		if es.IsDir() && es.Name()[0] != '.' {
			if !unicode.IsUpper([]rune(es.Name())[0]) {
				panic("failed: 文件夹 " + es.Name() + " 首字母必须为大写")
			}
		}
	}
	if *c {
		return
	}
	var files []string
	fs.WalkDir(os.DirFS("./"), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.Contains(path, "/") && unicode.IsUpper([]rune(path)[0]) {
			files = append(files, path)
		}
		return nil
	})
	fmt.Println(files)
	md5s := make([]string, len(files))
	for i, fn := range files {
		data, err := os.ReadFile(fn)
		if err != nil {
			panic(err)
		}
		buf := md5.Sum(data)
		*(*unsafe.Pointer)(unsafe.Pointer(&md5s[i])) = unsafe.Pointer(&buf)
		*(*uintptr)(unsafe.Add(unsafe.Pointer(&md5s[i]), unsafe.Sizeof(uintptr(0)))) = uintptr(16)
	}
	r := registry.NewRegedit("reilia.fumiama.top:32664", "", "fumiama", os.Getenv("REILIA_SPS"))
	err = r.Connect()
	if err != nil {
		panic(err)
	}
	for i, fn := range files {
		for c := 0; c < 16; c++ {
			err = r.Set("data/"+fn, md5s[i])
			fmt.Println("set", "data/"+fn, "=", hex.EncodeToString(helper.StringToBytes(md5s[i])))
			if err == nil {
				break
			}
			if c >= 15 {
				panic("ERROR:" + err.Error() + ", max retry times exceeded")
			} else {
				fmt.Println("ERROR:", err, ", retry times:", c)
			}
			_ = r.Close()
			process.SleepAbout1sTo2s()
			err = r.Connect()
			if err != nil {
				panic(err)
			}
		}
	}
	_ = r.Close()
}
