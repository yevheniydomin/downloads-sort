package main

import (
	"log"
	"os"
	"path"
	"strings"
	"sync"
)

func main() {
	downloadsPath := getDownloadsPath()

	files, err := os.ReadDir(downloadsPath)
	if err != nil {
		log.Fatalln("Could not read dir", err)
	}

	errorChannel := make(chan error, len(files))

	var wg sync.WaitGroup
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		extension := getFileExtension(file.Name())

		if extension == "" {
			wg.Add(1)
			go moveFile(downloadsPath, file.Name(), "no-extensions", &wg, errorChannel)
			continue
		}

		wg.Add(1)
		go moveFile(downloadsPath, file.Name(), extension, &wg, errorChannel)
	}
	wg.Wait()
	close(errorChannel)

	for err := range errorChannel {
		log.Printf("Error: %v", err)
	}
}

func IsDirExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func getDownloadsPath() string {
	path := os.Getenv("DOWNLOADS_PATH")
	if path == "" {
		log.Fatal("Please set DOWNLOADS_PATH env variable in your system with path to your Downloads folder")
	}

	isValidPass, err := IsDirExists(path)

	if err != nil || !isValidPass {
		log.Fatal("Could not set path to Downloads folder", err)
	}
	return path
}

func createDirIfNotExist(downloadsPath string, dirPath string) (bool, error) {
	dedicatedDir := path.Join(downloadsPath, dirPath)
	isDirExist, err := IsDirExists(dedicatedDir)
	if err != nil {
		log.Fatal("Couldn't handle createDirIfNotExist()", err)
		return false, err
	}

	if !isDirExist {
		err := os.Mkdir(dedicatedDir, 0700)
		if err != nil {
			log.Fatal("Couldn't create a new dir", err)
			return false, err
		}
	}
	return true, err
}

func getFileExtension(fileName string) string {
	extension := ""
	splitFileName := strings.Split(fileName, ".")
	if len(splitFileName) < 2 {
		return extension
	}
	extension = splitFileName[len(splitFileName)-1]
	return extension
}

func moveFile(downloadsPath string, fileName string, disDirName string, wg *sync.WaitGroup, errorChannel chan error) {
	defer wg.Done()
	oldLocation := path.Join(downloadsPath, fileName)
	newLocation := path.Join(downloadsPath, disDirName, fileName)
	isDirExists, err := createDirIfNotExist(downloadsPath, disDirName)
	if err != nil && !isDirExists {
		log.Fatalln("Can not move file because it's dir doesn't exist", err)
		errorChannel <- err
		return
	}

	error := os.Rename(oldLocation, newLocation)
	if error != nil {
		log.Fatal(error)
		errorChannel <- err
		return
	}
	errorChannel <- err
	return
}
