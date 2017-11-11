// Package tail implements tailing a log file.
//
// tail provides a channel on which log lines will be sent as string messages.
// one line in the log file is one message on the channel
package tail

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/hpcloud/tail"
	"golang.org/x/sys/unix"
)

type RotateStyle int

const (
	// foo.log gets rotated to foo.log.1, new entries go to foo.log
	RotateStyleSyslog RotateStyle = iota
	// foo.log.OLDSTAMP gets closed, new entries go to foo.log.NEWSTAMP
	// NOT YET IMPLEMENTED
	RotateStyleTimestamp
)

type TailOptions struct {
	ReadFrom  string `long:"read_from" description:"Location in the file from which to start reading. Values: beginning, end, last. Last picks up where it left off, if the file has not been rotated, otherwise beginning. When --backfill is set, it will override this option=beginning" default:"last"`
	Stop      bool   `long:"stop" description:"Stop reading the file after reaching the end rather than continuing to tail. When --backfill is set, it will override this option=true"`
	Poll      bool   `long:"poll" description:"use poll instead of inotify to tail files"`
	StateFile string `long:"statefile" description:"File in which to store the last read position. Defaults to a file in /tmp named $logfile.leash.state. If tailing multiple files, default is forced."`
}

// Statefile mechanics when ReadFrom is 'last'
// missing statefile => ReadFrom = end
// empty statefile => ReadFrom = end
// permission denied => WARN and ReadFrom = end
// invalid location (aka logfile's been rotated) => ReadFrom = beginning

type Config struct {
	// Path to the log file to tail
	Paths []string
	// Type of log rotation we expect on this file
	Type RotateStyle
	// Tail specific options
	Options TailOptions
}

// State is what's stored in a statefile
type State struct {
	INode  uint64 // the inode
	Offset int64
}

// GetSampledEntries wraps GetEntries and returns a list of channels that
// provide sampled entries
func GetSampledEntries(ctx context.Context, conf Config, sampleRate uint) ([]chan string, error) {
	unsampledLinesChans, err := GetEntries(ctx, conf)
	if err != nil {
		return nil, err
	}
	if sampleRate == 1 {
		return unsampledLinesChans, nil
	}

	sampledLinesChans := make([]chan string, 0, len(unsampledLinesChans))

	for _, lines := range unsampledLinesChans {
		sampledLines := make(chan string)
		go func(pLines chan string) {
			defer close(sampledLines)
			for line := range pLines {
				if shouldDrop(sampleRate) {
					logrus.WithFields(logrus.Fields{
						"line":       line,
						"samplerate": sampleRate,
					}).Debug("Sampler says skip this line")
				} else {
					sampledLines <- line
				}
			}
		}(lines)
		sampledLinesChans = append(sampledLinesChans, sampledLines)
	}
	return sampledLinesChans, nil
}

// shouldDrop returns true if the line should be dropped
// false if it should be kept
// if sampleRate is 5,
// on average one out of every 5 calls should return false
func shouldDrop(rate uint) bool {
	return rand.Intn(int(rate)) != 0
}

// GetEntries sets up a list of channels that get one line at a time from each
// file down each channel.
func GetEntries(ctx context.Context, conf Config) ([]chan string, error) {
	if conf.Type != RotateStyleSyslog {
		return nil, errors.New("Only Syslog style rotation currently supported")
	}
	// expand any globs in the list of files so our list all represents real files
	var filenames []string
	for _, filePath := range conf.Paths {
		if filePath == "-" {
			filenames = append(filenames, filePath)
		} else {
			files, err := filepath.Glob(filePath)
			if err != nil {
				return nil, err
			}
			files = removeStateFiles(files, conf)
			filenames = append(filenames, files...)
		}
	}
	if len(filenames) == 0 {
		return nil, errors.New("After removing missing files and state files from the list, there are no files left to tail")
	}

	// make our lines channel list; we'll get one channel for each file
	linesChans := make([]chan string, 0, len(filenames))
	numFiles := len(filenames)
	for _, file := range filenames {
		var lines chan string
		if file == "-" {
			lines = tailStdIn(ctx)
		} else {
			stateFile := getStateFile(conf, file, numFiles)
			tailer, err := getTailer(conf, file, stateFile)
			if err != nil {
				return nil, err
			}
			lines = tailSingleFile(ctx, tailer, file, stateFile)
		}
		linesChans = append(linesChans, lines)
	}

	return linesChans, nil
}

// removeStateFiles goes through the list of files and removes any that appear
// to be statefiles to avoid .leash.state.leash.state.leash.state from appearing
// when you use an overly permissive glob
func removeStateFiles(files []string, conf Config) []string {
	newFiles := []string{}
	for _, file := range files {
		if file == conf.Options.StateFile {
			logrus.WithFields(logrus.Fields{
				"file": file,
			}).Debug("skipping tailing file because it is named the same as the statefile flag")
			continue
		}
		if strings.HasSuffix(file, ".leash.state") {
			logrus.WithFields(logrus.Fields{
				"file": file,
			}).Debug("skipping tailing file because the filename ends with .leash.state")
			continue
		}
		// great! it's not a state file. let's use it.
		newFiles = append(newFiles, file)
	}
	return newFiles
}

func tailSingleFile(ctx context.Context, tailer *tail.Tail, file string, stateFile string) chan string {
	lines := make(chan string)
	// TODO report some metric to indicate whether we're keeping up with the
	// front of the file, of if it's being written faster than we can send
	// events

	stateFh, err := os.OpenFile(stateFile, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"logfile":   file,
			"statefile": stateFile,
		}).Warn("Failed to open statefile for writing. File location will not be saved.")
	}

	ticker := time.NewTicker(time.Second)
	state := State{}
	go func() {
		for range ticker.C {
			updateStateFile(&state, tailer, file, stateFh)
		}
	}()

	go func() {
	ReadLines:
		for {
			select {
			case line, ok := <-tailer.Lines:
				if !ok {
					// tailer.Lines is closed
					break ReadLines
				}
				if line.Err != nil {
					// skip errored lines
					continue
				}
				lines <- line.Text
			case <-ctx.Done():
				// will only trigger when the context is cancelled
				break ReadLines
			}
		}
		close(lines)
		ticker.Stop()
		updateStateFile(&state, tailer, file, stateFh)
		stateFh.Close()
	}()
	return lines
}

// tailStdIn is a special case to tail STDIN without any of the
// fancy stuff that the tail module provides
func tailStdIn(ctx context.Context) chan string {
	lines := make(chan string)
	input := bufio.NewReader(os.Stdin)
	go func() {
		defer close(lines)
		for {
			// check for signal triggered exit
			select {
			case <-ctx.Done():
				return
			default:
			}
			line, partialLine, err := input.ReadLine()
			if err != nil {
				logrus.Debug("stdin is closed")
				// bail when STDIN closes
				return
			}
			var parts []string
			parts = append(parts, string(line))
			for partialLine {
				line, partialLine, _ = input.ReadLine()
				parts = append(parts, string(line))
			}
			lines <- strings.Join(parts, "")
		}
	}()
	return lines
}

// getStartLocation reads the state file and creates an appropriate start
// location.  See details at the top of this file on how the loc is chosen.
func getStartLocation(stateFile string, logfile string) *tail.SeekInfo {
	beginning := &tail.SeekInfo{}
	end := &tail.SeekInfo{0, 2}
	fh, err := os.Open(stateFile)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"starting at": "end", "error": err,
		}).Debug("getStartLocation failed to open the statefile")
		return end
	}
	defer fh.Close()
	// read the contents of the state file (JSON)
	content := make([]byte, 1024)
	bytesRead, err := fh.Read(content)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"starting at": "end", "error": err,
		}).Debug("getStartLocation failed to read the statefile contents")
		return end
	}
	content = content[:bytesRead]
	// decode the contents of the statefile
	state := State{}
	if err := json.Unmarshal(content, &state); err != nil {
		logrus.WithFields(logrus.Fields{
			"starting at": "end", "error": err,
		}).Debug("getStartLocation failed to json decode the statefile")
		return end
	}
	// get the details of the existing log file
	logStat := unix.Stat_t{}
	if err := unix.Stat(logfile, &logStat); err != nil {
		logrus.WithFields(logrus.Fields{
			"starting at": "end", "error": err,
		}).Debug("getStartLocation failed to get unix.stat() on the logfile")
		return end
	}
	// compare inode numbers of the last-seen and existing log files
	if state.INode != logStat.Ino {
		logrus.WithFields(logrus.Fields{
			"starting at": "beginning", "error": err,
		}).Debug("getStartLocation found a different inode number for the logfile")
		// file's been rotated
		return beginning
	}
	logrus.WithFields(logrus.Fields{
		"starting at": state.Offset,
	}).Debug("getStartLocation seeking to offset in logfile")
	// we're good; start reading from the remembered state
	return &tail.SeekInfo{
		Offset: state.Offset,
		Whence: 0,
	}
}

// getTailer configures the *tail.Tail correctly to begin actually tailing the
// specified file.
func getTailer(conf Config, file string, stateFile string) (*tail.Tail, error) {
	// tail a real file
	var loc *tail.SeekInfo // 0 value means start at beginning
	var reOpen, follow bool = true, true
	switch conf.Options.ReadFrom {
	case "start", "beginning":
		// 0 value for tail.SeekInfo means start at beginning
	case "end":
		loc = &tail.SeekInfo{
			Offset: 0,
			Whence: 2,
		}
	case "last":
		loc = getStartLocation(stateFile, file)
	default:
		errMsg := fmt.Sprintf("unknown option to --read_from: %s",
			conf.Options.ReadFrom)
		return nil, errors.New(errMsg)
	}
	if conf.Options.Stop {
		reOpen = false
		follow = false
	}
	tailConf := tail.Config{
		Location:  loc,
		ReOpen:    reOpen, // keep reading on rotation, aka tail -F
		MustExist: true,   // fail if log file doesn't exist
		Follow:    follow, // don't stop at EOF, aka tail -f
		Logger:    tail.DiscardingLogger,
		Poll:      conf.Options.Poll, // use poll instead of inotify
	}
	logrus.WithFields(logrus.Fields{
		"tailConf":  tailConf,
		"conf":      conf,
		"statefile": stateFile,
		"location":  loc,
	}).Debug("about to call tail.TailFile")
	return tail.TailFile(file, tailConf)
}

// getStateFile returns the filename to use to track honeytail state.
//
// If a --tail.statefile parameter is provided, we try to respect it.
// It might describe an existing file, an existing directory, or a new path.
//
// If tailing a single logfile, we will use the specified --tail.statefile:
// - if it points to an existing file, that statefile will be used directly
// - if it points to a new path, that path will be written to directly
// - if it points to an existing directory, the statefile will be placed inside
//   the directory (and the statefile's name will be derived from the logfile).
//
// If honeytail is asked to tail multiple files, we will only respect the
// third case, where --tail.statefile describes an existing directory.
//
// The default behavior (or if --tail.statefile isn't respected) will be to
// write to the system's $TMPDIR/ and write to a statefile (where the name will
// be derived from the logfile).
func getStateFile(conf Config, filename string, numFiles int) string {
	confStateFile := os.TempDir()
	if conf.Options.StateFile != "" {
		info, err := os.Stat(conf.Options.StateFile)
		if numFiles == 1 && (os.IsNotExist(err) || (err == nil && !info.IsDir())) {
			return conf.Options.StateFile
		} else if err == nil && info.IsDir() {
			// If the --tail.statefile is a directory, write statefile inside the specified directory
			confStateFile = conf.Options.StateFile
		} else {
			logrus.Debugf("Couldn't write to --tail.statefile=%s, writing honeytail state for %s to $TMPDIR (%s) instead.",
				conf.Options.StateFile, filename, confStateFile)
		}
	}

	stateFileName := strings.TrimSuffix(filepath.Base(filename), ".log") + ".leash.state"
	return filepath.Join(confStateFile, stateFileName)
}

// updateStateFile updates the state file once per second with the current
// values for the logfile's inode number and offset
func updateStateFile(state *State, t *tail.Tail, file string, stateFh *os.File) {
	logStat := unix.Stat_t{}
	unix.Stat(file, &logStat)
	currentPos, err := t.Tell()
	if err != nil {
		return
	}
	state.INode = logStat.Ino
	state.Offset = currentPos
	out, err := json.Marshal(state)
	if err != nil {
		return
	}
	stateFh.Truncate(0)
	out = append(out, '\n')
	stateFh.WriteAt(out, 0)
	stateFh.Sync()
}
