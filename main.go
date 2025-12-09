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

	badger "github.com/dgraph-io/badger/v4"
)

func fileWalker(badgerdb *badger.DB) filepath.WalkFunc {

	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Print(err)
			return nil
		}

		if !info.IsDir() && !strings.HasSuffix(path, ".fsutils-badger") {
			f, err := os.Open(path)
			if err != nil {
				fmt.Println(err)
				return nil
			}
			defer f.Close()
			filehash := sha256.New()
			if _, err := io.Copy(filehash, f); err != nil {
				log.Fatal(err)
			}
			//sha256Sum := fmt.Sprintf("%x", h.Sum(nil))
			filehashSha256 := filehash.Sum(nil)

			txn := badgerdb.NewTransaction(true)
			defer txn.Discard()

			item, err := txn.Get([]byte(filehashSha256))
			if err != nil {
				log.Printf("key not found %s", filehashSha256)
			}

			var valCopy []byte

			if item != nil {
				valCopy, err = item.ValueCopy(nil)
				if err != nil {
					return err
				}
				fmt.Printf("The existing value is: %s\n", valCopy)
			}

			// Use the transaction...
			err = txn.Set(filehashSha256, []byte(path))
			if err != nil {
				return err
			}
			// Commit the transaction and check for error.
			if err := txn.Commit(); err != nil {
				return err
			}

			//hashes[sha256Sum] = append(hashes[sha256Sum], path) // md5s[md5value] = path
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
			badgerdb, err := badger.Open(badger.DefaultOptions(dir + "/.fsutils.badger"))
			if err != nil {
				log.Fatal(err)
			}
			defer badgerdb.Close()
			err = filepath.Walk(dir, fileWalker(badgerdb))
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
