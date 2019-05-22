package main

import (
	"crypto/sha1"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
)

var (
	baseDirectory      = flag.String("dir", "./", "base directory")
	fileMask           = flag.String("mask", ".*", "file mask regexp")
	sizeOnlyComparison = flag.Bool("size-only", false, "size only comparison")
)

func main() {
	// инициализация
	flag.Parse()
	fmt.Printf("DST - Duplicate Search Tool\nDirectory: %s\nFile mask: %s\n\n", *baseDirectory, *fileMask)

	// компилируем regexp
	fileMaskRegExp, err := regexp.Compile(*fileMask)
	if err != nil {
		fmt.Printf("[ERROR] Can't compile mask regexp [%s]\n", err.Error())
		return
	}

	// перебираем файлы в директории/поддерикториях, сохраянем в мапе [размер файла -> список файлов]
	fmt.Printf("Files iteration process started...\n")
	files := make(map[int64][]string)
	err = filepath.Walk(*baseDirectory, func(path string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !fileInfo.IsDir() {
			if fileMaskRegExp.FindString(fileInfo.Name()) != "" {
				var size = fileInfo.Size()
				files[size] = append(files[size], path)
			}
		}

		return nil
	})

	if err != nil {
		fmt.Printf("[ERROR] Can't walk base directory [%s]\n", err.Error())
		return
	}

	if *sizeOnlyComparison {
		// вывод файлов с одинаковым размером, по группам
		fmt.Printf("\n===================\nSome size files:\n===================\n\n")
		for _, list := range files {
			if len(list) > 1 {
				for _, file := range list {
					fmt.Printf("%s\n", file)
				}
				fmt.Println()
			}
		}
	} else {
		// хешируем все файлы с одинаковым размером, чтобы найти дубликаты
		fmt.Printf("Files comparison process started...\n")
		hashedFiles := make(map[string][]string)
		for _, list := range files {
			if len(list) > 1 {
				for _, file := range list {
					//fmt.Printf("[%s] hash calculation in progress...\n", file)
					hash, err := hashFile(file)
					if err != nil {
						fmt.Printf("[WARNING] Can't hash [%s] file [%s]\n", file, err.Error())
					} else {
						hashedFiles[hash] = append(hashedFiles[hash], file)
					}
				}
			}
		}

		// вывод одинковых файлов, по группам
		fmt.Printf("\n===================\nDuplicate files:\n===================\n\n")
		for _, files := range hashedFiles {
			if len(files) > 1 {
				for _, file := range files {
					fmt.Printf("%s\n", file)
				}
				fmt.Println()
			}
		}
	}
}

func hashFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = file.Close()
	}()

	hash := sha1.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
