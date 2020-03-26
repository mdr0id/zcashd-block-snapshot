package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"syscall"
	"time"

	"github.com/go-cmd/cmd"
)

var stopHeight int

func printUsage() {
	log.Fatalf("Usage: %s HEIGHT ", os.Args[0])
}

func main() {
	stopHeight := flag.Int("stop-height", 0, "What block height to stop at")
	exportDir := flag.String("export-dir", "./", "Where to write collected blocks")
	dataDir := flag.String("data-dir", "/srv/zcashd/.zcash", "Location of zcashd data files")
	flag.Parse()
	if *stopHeight == 0 {
		flag.PrintDefaults()
		log.Fatal("-stop-height but be a positive integer")
	}
	fmt.Printf("stopHeight: %d, dataDir: %s, exportDir: %s\n", *stopHeight, *dataDir, *exportDir)
	if err := testWriteToOutput(exportDir); err != nil {
		log.Fatalf("Exiting, can't write to output dir: %s, %s", *exportDir, err)
	}

	var err error
	if err != nil {
		fmt.Println("HEIGHT must be an integer")
		log.Fatalf("Usage: %s HEIGHT", os.Args[0])
	}
	fmt.Printf("Starting zcashd, stopping at height: %d\n", *stopHeight)
	cmdOptions := cmd.Options{
		Buffered:  false,
		Streaming: true,
	}

	// Create Cmd with options
	zcashdCmd := cmd.NewCmdOptions(cmdOptions, "zcashd", "-printtoconsole")

	// Print STDOUT and STDERR lines streaming from Cmd
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	doneChan := make(chan struct{})
	go func() {
		defer close(doneChan)
		// Done when both channels have been closed
		// https://dave.cheney.net/2013/04/30/curious-channels
		for zcashdCmd.Stdout != nil || zcashdCmd.Stderr != nil {
			select {
			case line, open := <-zcashdCmd.Stdout:
				if !open {
					zcashdCmd.Stdout = nil
					continue
				}
				height, err := reachedHeight(line)
				if err != nil {
					log.Fatal(err)
				}
				if height == nil {
					continue
				}
				if *height > *stopHeight {
					fmt.Printf("Somehow we passed the start height, stopping. At: %d, Want: %d\n", *height, *stopHeight)
					zcashdCmd.Stop()
					return
				}
				if *height == *stopHeight {
					fmt.Println("================== REACHED END HEIGHT ==================")
					zcashdCmd.Stop()
					fmt.Println("================== ZIPPING BLOCKS ==================")
					if err := gzipDefaultBlocks(*stopHeight, *dataDir, *exportDir); err != nil {
						log.Fatal(err)
					}
					return
				}
				if *height < *stopHeight {
					fmt.Printf("Updated height: %d, stopping at: %d\n", *height, *stopHeight)
					continue
				}
			case line, open := <-zcashdCmd.Stderr:
				if !open {
					zcashdCmd.Stderr = nil
					continue
				}
				fmt.Fprintln(os.Stderr, line)
			}
		}
	}()

	go func() {
		<-c
		zcashdCmd.Stop()
		os.Exit(1)
	}()

	// Run and wait for Cmd to return, discard Status
	<-zcashdCmd.Start()

	// Wait for goroutine to print everything
	<-doneChan
}

func reachedHeight(line string) (*int, error) {
	updateTipRE := regexp.MustCompile(`^UpdateTip`)
	if updateTipRE.Match([]byte(line)) {
		updateTipFullRE := regexp.MustCompile(`^UpdateTip:\s+new\s+best=(?P<hash>[[:xdigit:]]+)\s+height=(?P<height>[[:digit:]]+)[[:space:]]+bits=(?P<bits>[[:digit:]]+)[[:space:]]+log2_work=(?P<log2_work>[[:digit:].]+)[[:space:]]+tx=(?P<tx>[[:digit:]]+)[[:space:]]+date=(?P<date>[0-9-: ]+)progress=(?P<progress>[[:digit:].]+)[[:space:]]+cache=`)
		if updateTipFullRE.Match([]byte(line)) {
			names := updateTipFullRE.SubexpNames()
			result := updateTipFullRE.FindAllStringSubmatch(line, -1)[0]
			values := map[string]string{}
			for i, n := range result {
				values[names[i]] = n
			}
			fmt.Printf("New tip hash: %s Height: %s\n", values["hash"], values["height"])
			height, err := strconv.Atoi(values["height"])
			if err != nil {
				return nil, err
			}
			return &height, nil
		}
		log.Fatalf("Failed to process line: %s\n", line)
		return nil, errors.New("UpdateTip regex match failed")

	}
	fmt.Println(time.Now().Format("2006-01-02 15:04:05.000000"), " -- ", line)
	return nil, nil
}

func testWriteToOutput(outputDir *string) error {
	info, err := os.Stat(*outputDir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return errors.New("outputDir isn't a directory")
	}
	name := *outputDir + "/test-file"
	tFile, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	defer os.Remove(tFile.Name())
	return nil
}
