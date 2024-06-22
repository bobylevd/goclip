package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type FFMpegPlugin struct {
	fileQueue    *FileWatcherPlugin
	outputFolder string
	ffmpegFlags  []string
}

func NewFFMpegPlugin(fileQueue *FileWatcherPlugin, config Config) *FFMpegPlugin {
	return &FFMpegPlugin{
		fileQueue:    fileQueue,
		outputFolder: config.OutputFolder,
		ffmpegFlags:  config.FFMpegFlags,
	}
}

func (p *FFMpegPlugin) processFile(filename string) {
	if _, err := os.Stat(p.outputFolder); os.IsNotExist(err) {
		err := os.MkdirAll(p.outputFolder, os.ModePerm)
		if err != nil {
			log.Fatalf("Error creating output directory: %v", err)
		}
	}

	out := filepath.Join(p.outputFolder, filepath.Base(filename))

	var ffmpegCommand []string
	for _, flag := range p.ffmpegFlags {
		ffmpegCommand = append(ffmpegCommand, strings.Fields(flag)...)
	}
	ffmpegCommand = append([]string{"-i", filename}, ffmpegCommand...)
	ffmpegCommand = append(ffmpegCommand, out)

	log.Debugf("Processing file %s with ffmpeg", filename)

	cmd := exec.Command("ffmpeg", ffmpegCommand...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		log.Errorf("Failed to process file %s: %v, ffmpeg stderr: %s", filename, err, stderr.String())
	} else {
		log.Debugf("File %s was processed successfully", out)
		p.fileQueue.MarkProcessed(out)
	}
}

func (p *FFMpegPlugin) Run(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case file := <-p.fileQueue.toProcess:
			p.processFile(file)
		default:
			time.Sleep(1 * time.Second)
		}
	}
}
