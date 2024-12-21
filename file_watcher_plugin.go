package main

import (
	"path/filepath"
	"sync"
	"time"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
)

type FileWatcherPlugin struct {
	toProcess      chan string
	processedFiles chan string
	watchFolder    string
	outputFolder   string
	watcher        *fsnotify.Watcher
}

func NewFileWatcherPlugin(config Config) *FileWatcherPlugin {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logrus.Fatalf("Failed to create watcher: %v", err)
	}

	return &FileWatcherPlugin{
		toProcess:      make(chan string, 100),
		processedFiles: make(chan string, 100),
		watchFolder:    config.InputFolder,
		outputFolder:   config.OutputFolder,
		watcher:        watcher,
	}
}

func (fwp *FileWatcherPlugin) AddFile(filename string) {
	log.Debugf("Add file %s to be processed queue", filename)
	fwp.toProcess <- filename
}

func (fwp *FileWatcherPlugin) MarkProcessed(processedFile string) {
	log.Debugf("File %s was processed", processedFile)
	fwp.processedFiles <- processedFile
}

func (fwp *FileWatcherPlugin) GetNextProcessedFile() string {
	return <-fwp.processedFiles
}

func (fwp *FileWatcherPlugin) GetFileToProcess() string {
	return <-fwp.toProcess
}

func (fwp *FileWatcherPlugin) handleEvent(event fsnotify.Event) {
	filename := filepath.Base(event.Name)
	eventDir := filepath.Dir(event.Name)
	outputDir := filepath.Clean(fwp.outputFolder)
	watchDir := filepath.Clean(fwp.watchFolder)

	log.Debugf("Event received: Name=%s, Op=%s", event.Name, event.Op)

	// Ignore events in the output folder
	if eventDir == outputDir {
		log.Debugf("Ignored file in output folder: `%s`", event.Name)
		return
	}

	// Process files only in the watch folder
	if eventDir == watchDir {
		if !event.Op.Has(fsnotify.Remove) && filepath.Ext(filename) == ".mp4" && filename != "out.mp4" {
			log.Debugf("New file added: `%s`", event.Name)

			time.Sleep(2 * time.Second)

			if _, err := os.Stat(event.Name); err == nil {
                fwp.AddFile(event.Name)
            } else {
                log.Warnf("File not accessible after delay: `%s`: %v", event.Name, err)
            }
		} else {
			log.Debugf("Ignored file: `%s`", event.Name)
		}
	} else {
		log.Debugf("Ignored file outside of watch folder: `%s`", event.Name)
	}
}

func (fwp *FileWatcherPlugin) startMonitoring() {
	log.Debugf("Starting folder monitoring for folder: `%s`", fwp.watchFolder)
	err := fwp.watcher.Add(fwp.watchFolder)
	if err != nil {
		log.Fatalf("Failed to add watcher: %v", err)
	}

	for {
		select {
		case event, ok := <-fwp.watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Create == fsnotify.Create {
				fwp.handleEvent(event)
			}
		case err, ok := <-fwp.watcher.Errors:
			if !ok {
				return
			}
			log.Errorf("Error: %v", err)
		}
	}
}

func (fwp *FileWatcherPlugin) stopMonitoring() {
	log.Debugf("Stopped folder monitoring")
	err := fwp.watcher.Close()
	if err != nil {
		log.Fatalf("Failed to close watcher: %v", err)
	}
}

func (fwp *FileWatcherPlugin) Run(wg *sync.WaitGroup) {
	defer wg.Done()
	go fwp.startMonitoring()
	for {
		time.Sleep(1 * time.Second)
	}
}
