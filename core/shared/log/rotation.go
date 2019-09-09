package log

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	backupTimeFormat   = "2006-01-02T15-04-05.000"
	defaultFileMaxSize = 100
	megaByte           = 1024 * 1024
)

const (
	constRotateHour = iota
	constRotateFileSize
)

var (
	currentTime = time.Now
	osStat      = os.Stat
)

// Rotator the log-rotate impl
type Rotator struct {
	// file directory
	Dir string `json:"dir" yaml:"dir" xml:"dir"`
	// filename
	FileName string `json:"filename" yaml:"file_name" xml:"filename"`
	// max size of single row
	LogMaxSize int64 `json:"log_max_size" yaml:"log_max_size" xml:"log_max_size"`
	// rotate type
	RotateType int64 `json:"rotate_type" yaml:"rotate_type" xml:"rotate_type"`
	// max size of file, only be enabled in constRotateFileSize mode
	FileMaxSize int64 `json:"maxsize" yaml:"file_max_size" xml:"file_max_size"`
	// file expire time
	// constRotateFileSize mode: create time
	// constRotateHour mode: filename + create time
	MaxAge int64 `json:"max_age" yaml:"max_age" xml:"max_age"`
	// max backup count, only be enabled in constRotateFileSize mode
	MaxBackups int64 `json:"max_backups" yaml:"max_backups" xml:"max_backups"`
	// local time or not
	LocalTime bool `json:"localtime" yaml:"localtime" xml:"localtime"`

	prefix         string
	ext            string
	size           int64
	file           *os.File
	nextRotateTime time.Time
	millCh         chan bool
	mu             sync.Mutex
	startMill      sync.Once
}

// NewRotatorByHour return a rotator instance by hour
func NewRotatorByHour(dir, filename string) *Rotator {
	rot := &Rotator{
		Dir:        dir,
		FileName:   filename,
		LogMaxSize: 10240,
		RotateType: constRotateHour,
		MaxAge:     28,
	}
	return rot
}

// Write implements io.Writer
func (rot *Rotator) Write(p []byte) (n int, err error) {
	maxLen := rot.logMax()
	writeLen := int64(len(p))
	if writeLen > maxLen {
		return 0, fmt.Errorf("write length %d exceeds maximum record size %d", writeLen, maxLen)
	}

	rot.mu.Lock()
	defer rot.mu.Unlock()
	if rot.file == nil {
		if err = rot.openExistingOrNew(int64(len(p))); err != nil {
			return 0, err
		}
	}

	if rot.needRotate(int64(len(p))) {
		if err := rot.rotate(); err != nil {
			return 0, err
		}
	}

	n, err = rot.file.Write(p)
	rot.size += int64(n)
	return n, err
}

// Close implements io.Closer, and close the current file
func (rot *Rotator) Close() error {
	rot.mu.Lock()
	defer rot.mu.Unlock()
	return rot.close()
}

// GetFilePrefix get the file prefix
func (rot *Rotator) GetFilePrefix() string {
	return rot.prefix
}

// GetFileExt get the file ext
func (rot *Rotator) GetFileExt() string {
	return rot.ext
}

// GetRotateType get the rotation type
func (rot *Rotator) GetRotateType() int64 {
	return rot.RotateType
}

// internal func
func (rot *Rotator) isLogDir(d os.FileInfo) bool {
	if !d.IsDir() {
		return false
	}

	_, err := strconv.ParseInt(d.Name(), 10, 64)
	if err != nil && len(d.Name()) != 8 {
		return false
	}
	return true
}

func (rot *Rotator) isLogFile(f os.FileInfo) bool {
	if f.IsDir() {
		return false
	}

	prefix := rot.GetFilePrefix()
	ext := rot.GetFileExt()
	fileName := f.Name()

	if !strings.HasPrefix(fileName, prefix+"-") {
		return false
	}
	if !strings.HasSuffix(fileName, ext) {
		return false
	}
	return true
}

func (rot *Rotator) close() error {
	if rot.file == nil {
		return nil
	}
	err := rot.file.Close()
	rot.file = nil
	return err
}

func (rot *Rotator) needRotate(writeLen int64) bool {
	switch rot.RotateType {
	case constRotateHour:
		return time.Now().UnixNano() >= rot.nextRotateTime.UnixNano()
	default:
		return rot.size+writeLen > rot.fileMax()
	}
}

func (rot *Rotator) rotate() error {
	if err := rot.close(); err != nil {
		return err
	}
	if err := rot.openNew(); err != nil {
		return err
	}
	rot.mill()
	return nil
}

func (rot *Rotator) openNew() error {
	// dir
	err := os.MkdirAll(rot.dir(), 0744)
	if err != nil {
		return fmt.Errorf("can't make directories for new logfile: %s", err)
	}

	// file
	name := rot.filename()
	mode := os.FileMode(0644)
	info, err := osStat(name)
	if err == nil && rot.RotateType == constRotateFileSize && rot.MaxBackups > 0 {
		mode = info.Mode()
		newName := backupName(name, rot.LocalTime)
		if err := os.Rename(name, newName); err != nil {
			return fmt.Errorf("can't rename log file: %s", err)
		}

		if err := chown(name, info); err != nil {
			return err
		}
	}

	// create a new file
	f, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("can't open new logfile: %s", err)
	}
	rot.file = f
	rot.size = 0
	rot.nextRotateTime = getNextRotateTime()
	return nil
}

func (rot *Rotator) openExistingOrNew(writeLen int64) error {
	rot.mill()

	// get the base and ext name
	logFileName := filepath.Base(rot.baseFileName())
	rot.ext = filepath.Ext(logFileName)
	rot.prefix = logFileName[:len(logFileName)-len(rot.ext)]

	filename := rot.filename()
	info, err := osStat(filename)
	if os.IsNotExist(err) {
		return rot.openNew()
	}
	if err != nil {
		return fmt.Errorf("error getting log file info: %s", err)
	}

	if rot.RotateType != constRotateHour {
		if info.Size()+writeLen >= rot.fileMax() {
			return rot.rotate()
		}
	}

	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return rot.openNew()
	}
	rot.file = file
	rot.size = info.Size()
	rot.nextRotateTime = getNextRotateTime()
	return nil
}

func (rot *Rotator) millRunOnce() error {
	if rot.MaxBackups == 0 && rot.MaxAge == 0 {
		return nil
	}

	files, err := rot.oldLogFiles()
	if err != nil {
		return err
	}

	var remove []logInfo
	if rot.GetRotateType() == constRotateFileSize && rot.MaxBackups > 0 && rot.MaxBackups < int64(len(files)) {
		remove = append(remove, files[rot.MaxBackups:]...)
		files = files[0:rot.MaxBackups]
	}

	if rot.MaxAge > 0 {
		diff := time.Duration(int64(24*time.Hour) * int64(rot.MaxAge))
		cutoff := currentTime().Add(-1 * diff)

		var remaining []logInfo
		for _, f := range files {
			if f.timestamp.Before(cutoff) {
				remove = append(remove, f)
			} else {
				remaining = append(remaining, f)
			}
		}
		files = remaining
	}

	for _, f := range remove {
		errRemove := os.RemoveAll(filepath.Join(rot.Dir, f.Name()))
		if err == nil && errRemove != nil {
			err = errRemove
		}
	}

	return err
}

func (rot *Rotator) millRun() {
	for range rot.millCh {
		rot.millRunOnce()
	}
}

func (rot *Rotator) mill() {
	rot.startMill.Do(func() {
		rot.millCh = make(chan bool, 1)
		go rot.millRun()
	})
	select {
	case rot.millCh <- true:
	default:
	}
}

func (rot *Rotator) oldLogFiles() ([]logInfo, error) {
	files, err := ioutil.ReadDir(rot.Dir)
	if err != nil {
		return nil, fmt.Errorf("can't read log file directory: %s", err)
	}
	var logFiles []logInfo

	for _, f := range files {
		if rot.GetRotateType() == constRotateHour && rot.isLogDir(f) {
			logFiles = append(logFiles, logInfo{f.ModTime(), f})
			continue
		} else if rot.isLogFile(f) {
			logFiles = append(logFiles, logInfo{f.ModTime(), f})
			continue
		}
	}

	sort.Sort(byFormatTime(logFiles))
	return logFiles, nil
}

func (rot *Rotator) logMax() int64 {
	if rot.LogMaxSize != 0 {
		return int64(rot.LogMaxSize)
	}
	return int64(megaByte)

}

func (rot *Rotator) fileMax() int64 {
	if rot.FileMaxSize != 0 {
		return rot.FileMaxSize * int64(megaByte)
	}

	return int64(rot.FileMaxSize) * defaultFileMaxSize
}

func (rot *Rotator) filename() string {
	prefix, ext := rot.GetFilePrefix(), rot.GetFileExt()
	switch rot.RotateType {
	case constRotateHour:
		tm := currentTime()
		tmStr := fmt.Sprintf("%04d%02d%02d_%02d", tm.Year(), tm.Month(), tm.Day(), tm.Hour())
		dailyStr := fmt.Sprintf("%04d%02d%02d", tm.Year(), tm.Month(), tm.Day())
		return path.Join(rot.Dir, dailyStr, fmt.Sprintf("%s_%s%s", prefix, tmStr, ext))
	default:
		return path.Join(rot.Dir, rot.baseFileName())
	}
}

func (rot *Rotator) baseFileName() string {
	if rot.FileName == "" {
		return "rotate.log"
	}
	return rot.FileName
}

func (rot *Rotator) dir() string {
	switch rot.RotateType {
	case constRotateHour:
		tm := currentTime()
		suffix := fmt.Sprintf("%04d%02d%02d", tm.Year(), tm.Month(), tm.Day())
		return path.Join(rot.Dir, suffix)
	default:
		return rot.Dir
	}
}

func backupName(name string, local bool) string {
	dir := filepath.Dir(name)
	filename := filepath.Base(name)
	ext := filepath.Ext(filename)
	prefix := filename[:len(filename)-len(ext)]
	t := currentTime()
	if !local {
		t = t.UTC()
	}

	timestamp := t.Format(backupTimeFormat)
	return filepath.Join(dir, fmt.Sprintf("%s-%s%s", prefix, timestamp, ext))
}

func getNextRotateTime() time.Time {
	tm := currentTime()
	currHour := time.Date(tm.Year(), tm.Month(), tm.Day(), tm.Hour(), 0, 0, 0, time.Local)
	nextHour := currHour.Add(time.Hour)
	return nextHour
}

// logInfo use to delete old log files
type logInfo struct {
	timestamp time.Time
	os.FileInfo
}

type byFormatTime []logInfo

func (b byFormatTime) Less(i, j int) bool {
	return b[i].timestamp.After(b[j].timestamp)
}

func (b byFormatTime) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b byFormatTime) Len() int {
	return len(b)
}
