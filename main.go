package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func fileWalker(hashes map[string][]string) filepath.WalkFunc {

	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Print(err)
			return nil
		}

		if !info.IsDir() && !strings.HasSuffix(path, "sock") {
			f, err := os.Open(path)
			if err != nil {
				fmt.Println(err)
				return nil
			}
			defer f.Close()
			h := sha256.New()
			if _, err := io.Copy(h, f); err != nil {
				log.Fatal(err)
			}
			sha256Sum := fmt.Sprintf("%x", h.Sum(nil))
			hashes[sha256Sum] = append(hashes[sha256Sum], path) // md5s[md5value] = path
			// fmt.Printf("%s \t-> %s \n", path, md5value)
		}
		return nil
	}
}

func filtraHashesConPiuDiUnFile(hashes map[string][]string) map[string][]string {
	result := make(map[string][]string)
	for key, value := range hashes {
		if len(value) > 1 {
			result[key] = value
		}
	}
	return result
}
func usage() {
	fmt.Printf("fsutils usage: \n")
	fmt.Printf(" fsutils <cmd> <params> \n")
	fmt.Printf(" where cmd: \n")
	fmt.Printf("   - search-duplicates <DIR>")

}
func main() {
	log.SetFlags(log.Lshortfile)

	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}
	hashes := make(map[string][]string)

	cmd := os.Args[1]
	switch cmd {
	case "find-duplicates":
		{
			var dir string
			if len(os.Args) == 3 {
				dir = os.Args[2]
			} else {
				dir = "."
			}
			err := filepath.Walk(dir, fileWalker(hashes))
			if err != nil {
				log.Fatal(err)
			}

			b, err := json.MarshalIndent(filtraHashesConPiuDiUnFile(hashes), "", "  ")
			if err != nil {
				fmt.Println("error:", err)
			}
			os.Stdout.Write(b)
			os.Stdout.Write([]byte("\n"))

		}
	default:
		fmt.Printf("unknown command %s", cmd)
	}

	//fmt.Println("Files examined =", totFiles)
}
