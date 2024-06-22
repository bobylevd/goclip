package main

import (
	"github.com/go-yaml/yaml"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/sirupsen/logrus"
)

type Config struct {
	ChatID       int64    `yaml:"chat_id"`
	BotToken     string   `yaml:"bot_token"`
	InputFolder  string   `yaml:"input_folder"`
	OutputFolder string   `yaml:"output_folder"`
	FFMpegFlags  []string `yaml:"ffmpeg_flags"`
}

type Plugin interface {
	Run(wg *sync.WaitGroup)
}

type Clips struct {
	plugins []Plugin
}

func (sc *Clips) RegisterPlugin(plugin Plugin) {
	sc.plugins = append(sc.plugins, plugin)
}

func (sc *Clips) Run() {
	var wg sync.WaitGroup
	for _, plugin := range sc.plugins {
		wg.Add(1)
		go plugin.Run(&wg)
	}
	wg.Wait()
}

func main() {
	execPath, err := os.Executable()
	if err != nil {
		log.Fatalf("Failed to get executable path: %v", err)
	}
	execDir := filepath.Dir(execPath)

	configFilePath := filepath.Join(execDir, "config.yml")

	configFile, err := os.Open(configFilePath)
	if err != nil {
		log.Fatalf("Failed to open config file: %v", err)
	}
	defer configFile.Close()

	configData, err := io.ReadAll(configFile)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	var config Config
	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		log.Fatalf("Failed to unmarshal config data: %v", err)
	}

	logrus.Debug("Starting the application with the following configuration: ", config)

	app := &Clips{}

	fileQueuePlugin := NewFileWatcherPlugin(config)
	ffmpegPlugin := NewFFMpegPlugin(fileQueuePlugin, config)
	telegramPlugin := NewTelegramPlugin(fileQueuePlugin, config)

	app.RegisterPlugin(fileQueuePlugin)
	app.RegisterPlugin(ffmpegPlugin)
	app.RegisterPlugin(telegramPlugin)

	app.Run()
}
