package main

import (
	"bufio"
	"log"
	"os/exec"
)

func realtimeRead() {
	// The command you want to run along with the argument
	cmd := exec.Command("ping", "192.168.0.1000", "-c10")
	//fmt.Println("CMD STR", cmd.String())

	// Get a pipe to read from standard out
	rOut, _ := cmd.StdoutPipe()
	rErr, _ := cmd.StderrPipe()

	// Use the same pipe for standard error (if want to use 1 scanner/ pipe for both)
	//cmd.Stderr = cmd.Stdout

	// Make a new channel which will be used to ensure we get all output
	done := make(chan struct{})

	// Create a scanner which scans r in a line-by-line fashion
	outSc := bufio.NewScanner(rOut)
	errSc := bufio.NewScanner(rErr)

	// Use the scanner to scan the output line by line and log
	// if it's running in a goroutine so that it doesn't block
	go func() {

		// Read line by line and process it
		for outSc.Scan() {
			line := outSc.Text()
			log.Printf(line)
		}

		// We're all done, unblock the channel
		done <- struct{}{}

	}()
	go func() {

		// Read line by line and process it
		for errSc.Scan() {
			line := errSc.Text()
			log.Printf(line)
		}

		// We're all done, unblock the channel
		done <- struct{}{}

	}()

	// Start the command and check for errors
	err := cmd.Start()
	log.Println("ERR ", err)

	// Wait for all output to be processed
	<-done
	<-done

	// Wait for the command to finish
	err = cmd.Wait()
	log.Println(err)

}
