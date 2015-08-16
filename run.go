package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

var (
	lastMod time.Time

	errModifiedApp = errors.New("app has been modified")
)

func run() {
	if !isVolatile() {
		fmt.Println("volatile run: no runnable app detected")
		os.Exit(1)
	}

	appName, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	appName = filepath.Base(appName)

	// Prepare building
	buildCmd := exec.Command("go", "build", "-o", appName)
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr

	// Prepare running
	runCmd := exec.Command("./" + appName)
	runCmd.Stdout = os.Stdout
	runCmd.Stderr = os.Stderr

	running := false

	// Build and run
	if err := buildCmd.Run(); err == nil {
		runCmd.Start()
		running = true
	}

	// Modification detection
	lastMod = time.Now()
ModDetect:
	for {
		if err := filepath.Walk(".", modDetectWalk); err != nil {
			if err == errModifiedApp {
				break ModDetect
			}
			panic(err)
		}
		time.Sleep(500 * time.Millisecond)
	}

	// Interrupt
	if running {
		if err := runCmd.Process.Signal(os.Interrupt); err != nil {
			runCmd.Process.Kill()
		}
	}

	// Rerun
	log.Print("Rerunning server…\n\n")
	run()
}

func modDetectWalk(path string, fi os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if fi.ModTime().After(lastMod) {
		return errModifiedApp
	}
	return nil
}
