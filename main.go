package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

type FileInfo struct {
	Path string
	Size int64
	Hash string
}

var (
	db      *sql.DB
	dbMutex sync.Mutex
)

func main() {
	// CLI Flags
	dirPath := flag.String("path", ".", "Directory to scan")
	dbPath := flag.String("db", "deduplicate.db", "Path to SQLite database")
	flag.Usage = func() {
		fmt.Println("Usage: deduplicate [OPTIONS]")
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
		fmt.Println("\nExample:")
		fmt.Println("  deduplicate -path /your/folder -db results.db")
	}
	flag.Parse()

	if *dirPath == "" {
		fmt.Println("Error: -path is required")
		println()
		flag.Usage()
		os.Exit(1)
	}

	var err error
	db, err = sql.Open("sqlite3", *dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	createTable()

	var wg sync.WaitGroup
	fileChan := make(chan string)

	// Start worker goroutines
	for i := 0; i < 10; i++ {
		go worker(fileChan, &wg)
	}

	// Walk directory and send files to channel
	err = filepath.Walk(*dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		wg.Add(1)
		fileChan <- path
		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	wg.Wait()
	close(fileChan)
}

func worker(fileChan chan string, wg *sync.WaitGroup) {
	for path := range fileChan {
		processFile(path)
		wg.Done()
	}
}

func processFile(path string) {
	hash, size, err := hashFile(path)
	if err != nil {
		log.Printf("Error hashing file %s: %v\n", path, err)
		return
	}

	existingPath := checkDuplicate(hash, size)
	if existingPath != "" {
		handleDuplicate(path, existingPath)
		return
	}

	saveFile(FileInfo{Path: path, Size: size, Hash: hash})
}

func hashFile(path string) (string, int64, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", 0, err
	}
	defer file.Close()

	hash := sha256.New()
	size, err := io.Copy(hash, file)
	if err != nil {
		return "", 0, err
	}

	return hex.EncodeToString(hash.Sum(nil)), size, nil
}

func createTable() {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS files (
		id INTEGER PRIMARY KEY,
		path TEXT UNIQUE,
		size INTEGER,
		hash TEXT
	)`)
	if err != nil {
		log.Fatal(err)
	}
}

func checkDuplicate(hash string, size int64) string {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	var existingPath string
	err := db.QueryRow("SELECT path FROM files WHERE hash = ? AND size = ?", hash, size).Scan(&existingPath)
	if err == sql.ErrNoRows {
		return ""
	} else if err != nil {
		log.Fatal(err)
	}
	return existingPath
}

func handleDuplicate(duplicatePath, originalPath string) {
	dupDir := filepath.Join(filepath.Dir(duplicatePath), "duplicate_to_be_deleted")
	if err := os.MkdirAll(dupDir, os.ModePerm); err != nil {
		log.Printf("Failed to create duplicate directory: %v", err)
		return
	}

	newPath := filepath.Join(dupDir, filepath.Base(duplicatePath))
	if err := os.Rename(duplicatePath, newPath); err != nil {
		log.Printf("Failed to move duplicate file: %v", err)
		return
	}

	log.Printf("Duplicate found: %s (original: %s) -> moved to %s", duplicatePath, originalPath, newPath)
}

func saveFile(info FileInfo) {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	_, err := db.Exec("INSERT INTO files (path, size, hash) VALUES (?, ?, ?)", info.Path, info.Size, info.Hash)
	if err != nil {
		log.Printf("Failed to insert %s: %v", info.Path, err)
	}
}
