// Package tail implements tailing a log file.
//
// tail provides a channel on which log lines will be sent as string messages.
// one line in the log file is one message on the channel
package tail

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"
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
	ReadFrom  string `long:"read_from" description:"Location in the file from which to start reading. Values: beginning, end, last. Last picks up where it left off, if the file has not been rotated, otherwise beginning." default:"last"`
	Stop      bool   `long:"stop" description:"Stop reading the file after reaching the end rather than continuing to tail."`
	Poll      bool   `long:"poll" description:"use poll instead of inotify to tail files"`
	StateFile string `long:"statefile" description:"File in which to store the last read position. Defaults to a file with the same path as the log file and the suffix .leash.state. If tailing multiple files, default is forced."`
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

// GetSampledEntries wraps GetEntries and returns a channel that provides
// sampled entries
func GetSampledEntries(conf Config, sampleRate int) (chan string, error) {
	lines, err := GetEntries(conf)
	if err != nil {
		return nil, err
	}
	if sampleRate == 1 {
		return lines, nil
	}
	sampledLines := make(chan string)
	go func() {
		defer close(sampledLines)
		for line := range lines {
			if shouldSample(sampleRate) {
				sampledLines <- line
			} else {
				logrus.WithFields(logrus.Fields{
					"line": line,
				}).Debug("Sampler says skip this line")
			}
		}
	}()
	return sampledLines, nil
}

// shouldSample returns true if the line should be preserved
// false if it should be skipped
// if sampleRate is 5,
// on average one out of every 5 calls should return true
func shouldSample(sampleRate int) bool {
	if rand.Intn(sampleRate) == 0 {
		return true
	}
	return false
}

// GetEntries opens the log file, reading from the end. It sends one line
// at a time down the returned channel
func GetEntries(conf Config) (chan string, error) {
	if conf.Type != RotateStyleSyslog {
		return nil, errors.New("Only Syslog style rotation currently supported")
	}
	lines := make(chan string)
	var wg sync.WaitGroup
	defer func() {
		go func() {
			wg.Wait()
			close(lines)
		}()
	}()
	// handle reading from STDIN
	if conf.Paths[0] == "-" {
		return lines, tailStdIn(lines, &wg)
	}
	for _, filePath := range conf.Paths {
		if err := tailMultipleFiles(conf, filePath, lines, &wg); err != nil {
			return nil, err
		}
	}
	// close lines when all processors are done

	return lines, nil
}

func tailMultipleFiles(conf Config, filePath string, lines chan string, wg *sync.WaitGroup) error {
	files, err := filepath.Glob(filePath)
	if err != nil {
		return err
	}
	if len(files) > 1 {
		// when tailing multiple files, force the default statefile use
		conf.Options.StateFile = ""
	}
	for _, file := range files {
		var realStateFile string
		if conf.Options.StateFile == "" {
			baseName := strings.TrimSuffix(file, ".log")
			realStateFile = baseName + ".leash.state"
		} else {
			realStateFile = conf.Options.StateFile
		}
		if err := tailSingleFile(conf, file, realStateFile, lines, wg); err != nil {
			return err
		}
	}
	return nil
}

func tailSingleFile(conf Config, file string, stateFile string, lines chan string, wg *sync.WaitGroup) error {
	// TODO report some metric to indicate whether we're keeping up with the
	// front of the file, of if it's being written faster than we can send
	// events

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
		return errors.New(errMsg)
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
	t, err := tail.TailFile(file, tailConf)
	logrus.WithFields(logrus.Fields{"tail": t}).Debug("finished call to TailFile")
	if err != nil {
		return err
	}
	// TODO this only updates once/sec. On clean shutdown, make sure we write
	// one last time after stopping reading traffic.
	go updateStateFile(t, stateFile, file)
	wg.Add(1)
	go func() {
		for line := range t.Lines {
			if line.Err != nil {
				// skip errored lines
				continue
			}
			lines <- line.Text
		}
		wg.Done()
	}()
	return nil
}

// tailStdIn is a special case to tail STDIN without any of the
// fancy stuff that the tail module provides
func tailStdIn(lines chan string, wg *sync.WaitGroup) error {
	input := bufio.NewReader(os.Stdin)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
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
	return nil
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

// updateStateFile updates the state file once per second with the current
// values for the logfile's inode number and offset
func updateStateFile(t *tail.Tail, stateFile string, file string) {
	statefh, err := os.OpenFile(stateFile, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"logfile":   file,
			"statefile": stateFile,
		}).Warn("Failed to open statefile for writing. File location will not be saved.")
		return
	}
	ticker := time.NewTicker(time.Second)
	state := State{}
	for _ = range ticker.C {
		logStat := unix.Stat_t{}
		unix.Stat(file, &logStat)
		currentPos, err := t.Tell()
		if err != nil {
			continue
		}
		state.INode = logStat.Ino
		state.Offset = currentPos
		out, err := json.Marshal(state)
		if err != nil {
			continue
		}
		statefh.Truncate(0)
		out = append(out, '\n')
		statefh.WriteAt(out, 0)
		statefh.Sync()
	}
}
