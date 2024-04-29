package main

import (
	"fmt"
	"github.com/nfnt/resize"
	"image"
	"image/jpeg"
	"log"
	"os"
	"sync"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Fatalf("recovered from panic: %v", r)
		}
	}()

	log.Printf("starting glitch")

	dir := "/Users/eaardal/Pictures/photos/tmp"
	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
		return
	}

	var fileNames []string
	for _, entry := range entries {
		log.Printf("ENTRY: %+v", entry)
		if entry.IsDir() {
			log.Printf("skipping subdir %s", entry.Name())
			continue
		}
		if entry == nil {
			log.Printf("skipping nil entry")
			continue
		}
		if entry.Name() == "" {
			log.Printf("skipping entry with empty name")
			continue
		}
		fileNames = append(fileNames, entry.Name())
	}

	log.Printf("fileinfos: %+v", fileNames)

	if err := makeThumbnails(fileNames); err != nil {
		log.Fatal(err)
	}

	log.Printf("exiting glitch")
}

func waitclose(wg *sync.WaitGroup, done chan string, errch chan error) {
	log.Printf("waiting...")
	wg.Wait()
	log.Printf("done waiting")

	log.Printf("closing channels...")
	close(done)
	close(errch)
	log.Printf("closed channels")
}

func makeThumbnails(images []string) error {
	done := make(chan string)
	errs := make(chan error)

	var wg sync.WaitGroup
	wg.Add(len(images))

	go waitclose(&wg, done, errs)

	for _, img := range images {
		go makeThumbnail(img, &wg, done, errs)
	}

	for newName := range done {
		log.Printf("new thumb file: %s", newName)
	}
	for thumbErr := range errs {
		log.Printf("thumb err: %v", thumbErr)
	}

	return nil
}

func makeThumbnail(imgName string, wg *sync.WaitGroup, done chan string, errCh chan error) {
	defer func() {
		log.Printf("waitgroup done")
		wg.Done()
	}()

	log.Printf("opening image file %s", imgName)

	file, err := os.Open(imgName)
	if err != nil {
		errCh <- fmt.Errorf("open file %s: %w", imgName, err)
		done <- ""
		return
	}
	defer func() {
		log.Printf("defer close file")
		if err := file.Close(); err != nil {
			errCh <- fmt.Errorf("closing file: %w", err)
			done <- ""
			return
		}
	}()

	log.Printf("opened file %s", imgName)

	srcImg, _, err := image.Decode(file)
	if err != nil {
		errCh <- fmt.Errorf("decode image: %w", err)
		done <- ""
		return
	}

	log.Printf("decoded %s", file.Name())

	dstImage := resize.Resize(80, 80, srcImg, resize.Lanczos3)

	log.Printf("resized %s", file.Name())

	newName := fmt.Sprintf("%s.thumb.jpg", file.Name())
	newFile, err := os.Create(newName)
	if err != nil {
		errCh <- fmt.Errorf("create thumb file: %w", err)
		done <- ""
		return
	}

	log.Printf("created thumb file %s", newName)

	if err := jpeg.Encode(newFile, dstImage, nil); err != nil {
		errCh <- fmt.Errorf("encode jpeg: %w", err)
		done <- ""
		return
	}

	log.Printf("encoded new thumb image")
	done <- newName
}
