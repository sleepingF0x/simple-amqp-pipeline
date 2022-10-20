package main

import (
	"context"
	"flag"
	"log"
	"os"
	"path/filepath"
	"simple-amqp-pipeline/config"
	"simple-amqp-pipeline/pipeline"
	"simple-amqp-pipeline/version"
)

var (
	configDir = flag.String("path", "./conf.d", "AMQP pipeline config directory")
	showVer   = flag.Bool("version", false, "show version")
)

func init() {
	flag.Parse()
}

func main() {
	if *showVer {
		log.Println("version: ", version.Full())
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var configFiles []string
	_ = filepath.Walk(*configDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatalf("walk pipeline directory err : %s", err)
			return err
		}
		if !info.IsDir() {
			configFiles = append(configFiles, path)
		}
		return nil
	})

	for _, c := range configFiles {
		conf, err := config.LoadConfiguration(c)
		if err != nil {
			log.Printf("load pipeline config %v err : %v", c, err)
			continue
		}
		pipe, err := pipeline.NewPipeline(conf.SrcConf, conf.DstConf, conf.Workers)
		if err != nil {
			log.Println("create new pipeline failed")
		} else {
			pipe.Start()
		}
	}
	<-ctx.Done()
}
