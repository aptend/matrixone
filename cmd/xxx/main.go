package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"
)

func main() {
	// move()
	rollback()
}

func rollback() {
	targetDir := "/root/matrixone/mo-data/shared"
	dirname := "/root/matrixone/backup"
	files, err := ioutil.ReadDir(dirname)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		err := os.Rename(path.Join(dirname, file.Name()), path.Join(targetDir, file.Name()))
		if err != nil {
			log.Fatal(err)
		}
	}

}

func move() {
	dryRun := true
	name := "/root/matrixone/objectlist"
	dirname := "/root/matrixone/mo-data/shared"
	targetDir := "/root/matrixone/backup"

	f1, err := os.Open(name)
	defer f1.Close()

	pinned := make(map[string]bool)

	scanner := bufio.NewScanner(f1)
	for scanner.Scan() {
		// fmt.Println("111", scanner.Text())
		pinned[scanner.Text()] = true
	}

	// list all the files in the directory
	files, err := ioutil.ReadDir(dirname)
	if err != nil {
		log.Fatal(err)
	}
	toMove := make(map[string]bool)
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		modTime := file.ModTime()
		if time.Since(modTime).Minutes() < 10 {
			continue
		}
		fmt.Println("222", file.Name())
		if _, ok := pinned[file.Name()]; !ok {
			toMove[file.Name()] = true
		}
	}

	fmt.Printf("Total files to move: %d/%d\n", len(toMove), len(files))
	if dryRun {
		return
	}

	os.MkdirAll(targetDir, os.ModePerm)
	for file := range toMove {
		err := os.Rename(path.Join(dirname, file), path.Join(targetDir, file))
		if err != nil {
			log.Fatal(err)
		}
	}
}
