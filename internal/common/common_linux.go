// +build linux

package common

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync/atomic"
)

func DoSysctrl(mib string) ([]string, error) {
	err := os.Setenv("LC_ALL", "C")
	if err != nil {
		return []string{}, err
	}
	sysctl, err := exec.LookPath("/sbin/sysctl")
	if err != nil {
		return []string{}, err
	}
	out, err := exec.Command(sysctl, "-n", mib).Output()
	if err != nil {
		return []string{}, err
	}
	v := strings.Replace(string(out), "{ ", "", 1)
	v = strings.Replace(string(v), " }", "", 1)
	values := strings.Fields(string(v))

	return values, nil
}

func NumProcs() (uint64, error) {
	f, err := os.Open(HostProc())
	if err != nil {
		return 0, err
	}
	defer f.Close()

	list, err := f.Readdir(-1)
	if err != nil {
		return 0, err
	}
	return uint64(len(list)), err
}

// cachedBootTime must be accessed via atomic.Load/StoreUint64
var cachedBootTime uint64

// BootTime returns the system boot time expressed in seconds since the epoch.
func BootTime() (uint64, error) {
	t := atomic.LoadUint64(&cachedBootTime)
	if t != 0 {
		return t, nil
	}
	filename := HostProc("stat")
	lines, err := ReadLines(filename)
	if err != nil {
		return 0, err
	}
	for _, line := range lines {
		if strings.HasPrefix(line, "btime") {
			f := strings.Fields(line)
			if len(f) != 2 {
				return 0, fmt.Errorf("wrong btime format")
			}
			b, err := strconv.ParseInt(f[1], 10, 64)
			if err != nil {
				return 0, err
			}
			t = uint64(b)
			atomic.StoreUint64(&cachedBootTime, t)
			return t, nil
		}
	}

	return 0, fmt.Errorf("could not find btime")
}