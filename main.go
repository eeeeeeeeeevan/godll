package main

import (
	"flag"
	"path/filepath"
	"time"

	"golang.org/x/sys/windows"
)

var (
	executable string
	arguments  string
	library    string
	delay      float64
)

func init() {
	flag.StringVar(&executable, "executable", "", "target exec")
	flag.StringVar(&arguments, "arguments", "", "args")
	flag.StringVar(&library, "library", "", "dll path")
	flag.Float64Var(&delay, "delay", 0, "arb. delay")
}

func main() {
	flag.Parse()

	proc, err := procstart()
	if err != nil {
		panic(err)
	}

	if delay > 0 {
		time.Sleep(time.Duration(delay * float64(time.Second)))
	}

	dllinject(proc)
}

func procstart() (*windows.ProcessInformation, error) {
	exe, _ := windows.UTF16PtrFromString(executable)
	var args *uint16
	if arguments != "" {
		args, _ = windows.UTF16PtrFromString(arguments)
	}

	var si windows.StartupInfo
	var pi windows.ProcessInformation

	err := windows.CreateProcess(exe, args, nil, nil, false, 0, nil, nil, &si, &pi)

	return &pi, err
}

func dllinject(pi *windows.ProcessInformation) {
	k32 := windows.NewLazySystemDLL("kernel32.dll")
	loadLibrary := k32.NewProc("LoadLibraryA")

	dllpath, _ := filepath.Abs(library)
	pathbytes := append([]byte(dllpath), 0)

	process, _ := windows.OpenProcess(windows.PROCESS_ALL_ACCESS, false, pi.ProcessId)

	rbuf, _ := windows.VirtualAllocEx(process, 0, uintptr(len(pathbytes)), windows.MEM_COMMIT|windows.MEM_RESERVE, windows.PAGE_READWRITE)

	var written uintptr
	windows.WriteProcessMemory(process, rbuf, &pathbytes[0], uintptr(len(pathbytes)), &written)

	windows.CreateRemoteThread(process, nil, 0, loadLibrary.Addr(), rbuf, 0, nil)

	time.Sleep(time.Second)
	windows.VirtualFreeEx(process, rbuf, 0, windows.MEM_RELEASE)
	windows.CloseHandle(process)
}
