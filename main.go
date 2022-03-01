package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"unicode"
	"unsafe"

	"github.com/fumiama/go-registry"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"
)

func main() {
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
	r := registry.NewRegedit("reilia.westeurope.cloudapp.azure.com:32664", "fumiama", os.Getenv("REILIA_SPS"))
	err := r.Connect()
	if err != nil {
		panic(err)
	}
	for i, fn := range files {
		fmt.Println("set", "data/"+fn, "=", hex.EncodeToString(helper.StringToBytes(md5s[i])))
		err = r.Set("data/"+fn, md5s[i])
		if err != nil {
			panic(err)
		}
	}
	r.Close()
}
