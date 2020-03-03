package wasm

/*
#cgo CFLAGS: -Iinclude

#include "wasi_core.h"

*/
import "C"

import (
	"fmt"
	"strings"
	"unsafe"

	wasm3 "github.com/ltearno/go-wasm3"
)

var preopen = []string{
	"<stdin>",
	"<stdout>",
	"<stderr>",
	"./",
	"../",
}

const (
	WASI_ESUCCESS                  = C.__WASI_ESUCCESS
	WASI_EBADF                     = C.__WASI_EBADF
	WASI_PREOPENTYPE_DIR           = C.__WASI_PREOPENTYPE_DIR
	WASI_FILETYPE_CHARACTER_DEVICE = C.__WASI_FILETYPE_CHARACTER_DEVICE
	WASI_FILETYPE_DIRECTORY        = C.__WASI_FILETYPE_DIRECTORY
	WASI_FILETYPE_REGULAR_FILE     = C.__WASI_FILETYPE_REGULAR_FILE
)

func setWasiStat(fd int, fdStatAddr unsafe.Pointer) {
	fdStat := (*C.__wasi_fdstat_t)(fdStatAddr)
	if fd >= 0 && fd <= 2 {
		fdStat.fs_filetype = C.__WASI_FILETYPE_CHARACTER_DEVICE
	} else if fd == 3 || fd == 4 {
		fdStat.fs_filetype = C.__WASI_FILETYPE_DIRECTORY
	} else {
		fdStat.fs_filetype = C.__WASI_FILETYPE_REGULAR_FILE
	}
	fdStat.fs_flags = 0
	fdStat.fs_rights_base = C.ulong(0xfffffffffff)
	fdStat.fs_rights_inheriting = C.ulong(0xfffffffffff)
}

type state struct {
	wctx *WasmProcessContext

	Arguments []string

	WasiFilename         string
	CmdLineNbArgs        int
	CmdLineArgs          []string
	CmdLineArgsSize      int
	OpenedVirtualFiles   map[int]VirtualFile
	NextFileDescriptorID int

	WasiExitValue uint32
}

func NewWASIHostPlugin(wasiFileName *string, arguments *[]string, preopenedFiles map[int]VirtualFile) WASMAPIPlugin {
	s := &state{
		OpenedVirtualFiles:   make(map[int]VirtualFile),
		NextFileDescriptorID: 42,
	}

	if wasiFileName != nil {
		s.WasiFilename = *wasiFileName
	} else {
		s.WasiFilename = "a.out"
	}

	if arguments != nil {
		s.Arguments = *arguments
	} else {
		s.Arguments = []string{}
	}

	for k, v := range preopenedFiles {
		s.OpenedVirtualFiles[k] = v
	}

	return s
}

type WasiCallHandler func(state *state, cs *CallSite) (uint32, int)

func (state *state) BindAPIFunction(moduleName string, functionName string, signature string, handler WasiCallHandler) {
	state.wctx.Runtime.AttachFunction(moduleName, functionName, signature, func(runtime wasm3.RuntimeT, sp unsafe.Pointer, mem unsafe.Pointer) int {
		callSite := &CallSite{
			sp:  sp,
			mem: mem,
		}

		result, m3PossibleTrap := handler(state, callSite)

		if m3PossibleTrap == 0 {
			*(*uint32)(sp) = result
		}

		return m3PossibleTrap
	})
}

func (s *state) Bind(wctx *WasmProcessContext) {
	s.wctx = wctx

	importedModules := wctx.GetImportedModules()

	var wasiModuleName string
	if _, ok := importedModules["wasi_unstable"]; ok {
		// official temporary wasi
		wasiModuleName = "wasi_unstable"
	} else if _, ok := importedModules["wasi_snapshot_preview1"]; ok {
		wasiModuleName = "wasi_snapshot_preview1"
	} else {
		return
	}

	if wctx.Fctx.Trace {
		fmt.Println("preparing WASI host", s.WasiFilename, s.Arguments, "...")
	}

	s.prepareCmdLineArgs()

	s.BindAPIFunction(wasiModuleName, "path_open", "i(ii*iiiii*)", WASIPathOpen)
	s.BindAPIFunction(wasiModuleName, "proc_exit", "i(i)", WASIProcExit)
	s.BindAPIFunction(wasiModuleName, "fd_readdir", "i(i*ii*)", WASIFdReadDir)
	s.BindAPIFunction(wasiModuleName, "args_sizes_get", "i(**)", WASIArgsSizesGet)
	s.BindAPIFunction(wasiModuleName, "args_get", "i(**)", WASIArgsGet)
	s.BindAPIFunction(wasiModuleName, "fd_prestat_get", "i(i*)", WASIFdPrestatGet)
	s.BindAPIFunction(wasiModuleName, "fd_prestat_dir_name", "i(i*i)", WASIFdPrestatDirName)
	s.BindAPIFunction(wasiModuleName, "fd_fdstat_get", "i(i*)", WASIFdFdStatGet)
	s.BindAPIFunction(wasiModuleName, "path_filestat_get", "i(ii*i*)", WASIPathFilestatGet)
	s.BindAPIFunction(wasiModuleName, "fd_write", "i(iii*)", WASIFdWrite)
	s.BindAPIFunction(wasiModuleName, "fd_read", "i(iii*)", WASIFdRead)
	s.BindAPIFunction(wasiModuleName, "fd_seek", "i(iii*)", WASIFdSeek)
	s.BindAPIFunction(wasiModuleName, "fd_close", "i(i)", WASIFdClose)
	s.BindAPIFunction(wasiModuleName, "environ_sizes_get", "i(**)", WASIEnvironSizesGet)
	s.BindAPIFunction(wasiModuleName, "environ_get", "i(**)", WASIEnvironGet)

	wctx.BindNotYetImplementedFunction(wasiModuleName, "fd_fdstat_set_flags", "i(ii)")
	wctx.BindNotYetImplementedFunction(wasiModuleName, "fd_datasync", "i(i)")
	wctx.BindNotYetImplementedFunction(wasiModuleName, "clock_res_get", "i(i*)")
	wctx.BindNotYetImplementedFunction(wasiModuleName, "clock_time_get", "i(ii*)")
	wctx.BindNotYetImplementedFunction(wasiModuleName, "fd_renumber", "i(ii)")
	wctx.BindNotYetImplementedFunction(wasiModuleName, "fd_tell", "i(i*)")
	wctx.BindNotYetImplementedFunction(wasiModuleName, "fd_fdstat_set_rights", "i(iii)")
	wctx.BindNotYetImplementedFunction(wasiModuleName, "fd_advise", "i(iiii)")
	wctx.BindNotYetImplementedFunction(wasiModuleName, "fd_allocate", "i(iii)")
	wctx.BindNotYetImplementedFunction(wasiModuleName, "path_create_directory", "i(i*i)")
	wctx.BindNotYetImplementedFunction(wasiModuleName, "path_link", "i(ii*ii*i)")
	wctx.BindNotYetImplementedFunction(wasiModuleName, "path_readlink", "i(i*i*i*)")
	wctx.BindNotYetImplementedFunction(wasiModuleName, "path_rename", "i(i*ii*i)")
	wctx.BindNotYetImplementedFunction(wasiModuleName, "fd_filestat_get", "i(i*)")
	wctx.BindNotYetImplementedFunction(wasiModuleName, "fd_filestat_set_times", "i(iiii)")
	wctx.BindNotYetImplementedFunction(wasiModuleName, "fd_filestat_set_size", "i(ii)")
	wctx.BindNotYetImplementedFunction(wasiModuleName, "path_filestat_set_times", "i(ii*iiii)")
	wctx.BindNotYetImplementedFunction(wasiModuleName, "path_symlink", "i(*ii*i)")
	wctx.BindNotYetImplementedFunction(wasiModuleName, "path_unlink_file", "i(i*i)")
	wctx.BindNotYetImplementedFunction(wasiModuleName, "path_remove_directory", "i(i*i)")
	wctx.BindNotYetImplementedFunction(wasiModuleName, "poll_oneoff", "i(**i*)")
	wctx.BindNotYetImplementedFunction(wasiModuleName, "proc_raise", "i(i)")
	wctx.BindNotYetImplementedFunction(wasiModuleName, "random_get", "i(*i)")
	wctx.BindNotYetImplementedFunction(wasiModuleName, "sock_recv", "i(i*ii**)")
	wctx.BindNotYetImplementedFunction(wasiModuleName, "sock_send", "i(i*ii*)")
	wctx.BindNotYetImplementedFunction(wasiModuleName, "sock_shutdown", "i(ii)")
	wctx.BindNotYetImplementedFunction(wasiModuleName, "sched_yield", "i()")
}

func (s *state) prepareCmdLineArgs() {
	s.CmdLineNbArgs = 1 + len(s.Arguments)
	s.CmdLineArgs = make([]string, s.CmdLineNbArgs)
	s.CmdLineArgs[0] = s.WasiFilename
	for i, arg := range s.Arguments {
		s.CmdLineArgs[i+1] = arg
	}
	s.CmdLineArgsSize = 0
	for _, arg := range s.CmdLineArgs {
		s.CmdLineArgsSize = s.CmdLineArgsSize + len(arg)
	}
	s.CmdLineArgsSize = s.CmdLineArgsSize + s.CmdLineNbArgs
}

// WASIPathOpen is path_open
func WASIPathOpen(state *state, cs *CallSite) (uint32, int) {
	dirFd := cs.GetParamUINT32(0)
	lookupFlags := cs.GetParamUINT32(1)
	path := cs.GetParamString(2, 3)
	oFlags := cs.GetParamUINT32(4)
	rightsBase := cs.GetParamUINT32(5)
	rightsInheriting := cs.GetParamUINT32(6)
	flags := cs.GetParamUINT32(7)
	fd := cs.GetParamUINT32Ptr(8)

	if state.wctx.Fctx.Trace {
		fmt.Printf("called path_open dirFd %d lookupFlags %x oFlags %x rightsBase %x rightsInheriting %x flags %x fdAddr %x  '%s'\n",
			dirFd,
			lookupFlags,
			oFlags,
			rightsBase,
			rightsInheriting,
			flags,
			fd,
			path,
		)
	}

	var virtualFile VirtualFile = nil

	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		virtualFile = CreateWebAccessVirtualFile(path)
	} else if path == "api://input" {
		virtualFile = CreateInputVirtualFile(state.wctx)
	} else if path == "api://output" {
		virtualFile = CreateOutputVirtualFile(state.wctx)
	} else {
		*fd = 0
		return WASI_EBADF, 0
	}

	*fd = uint32(state.NextFileDescriptorID)
	state.NextFileDescriptorID = state.NextFileDescriptorID + 1
	state.OpenedVirtualFiles[int(*fd)] = virtualFile

	return WASI_ESUCCESS, 0
}

// WASIProcExit WASIProcExit
func WASIProcExit(state *state, cs *CallSite) (uint32, int) {
	exitValue := cs.GetParamUINT32(0)

	if state.wctx.Fctx.Trace {
		fmt.Printf("called proc_exit with value %d\n", exitValue)
	}

	state.WasiExitValue = exitValue
	state.wctx.Fctx.Result = int(exitValue)

	return WASI_ESUCCESS, 1
}

// WASIFdReadDir WASIFdReadDir
func WASIFdReadDir(state *state, cs *CallSite) (uint32, int) {
	fd := cs.GetParamUINT32(0)
	buffer := cs.GetParamByteBuffer(1, 2)
	cookie := cs.GetParamUINT32(3)
	bufusedAddr := cs.GetParamPointer(4)

	if state.wctx.Fctx.Trace {
		fmt.Printf("called fd_readdir fd %d bufAddr %x cookie %d bufusedAddr %x\n", fd, &buffer, cookie, bufusedAddr)
	}

	*(*uint32)(bufusedAddr) = 0

	return WASI_ESUCCESS, 0
}

// WASIArgsSizesGet WASIArgsSizesGet
func WASIArgsSizesGet(state *state, cs *CallSite) (uint32, int) {
	argc := cs.GetParamUINT32Ptr(0)
	argvBufSize := cs.GetParamUINT32Ptr(1)

	if state.wctx.Fctx.Trace {
		fmt.Printf("called args_sizes_get %x %x\n", argc, argvBufSize)
	}

	*argc = uint32(state.CmdLineNbArgs)
	*argvBufSize = uint32(state.CmdLineArgsSize)

	return WASI_ESUCCESS, 0
}

// WASIArgsGet WASIArgsGet
func WASIArgsGet(state *state, cs *CallSite) (uint32, int) {
	p0 := cs.GetParamUINT32(0)
	p1 := cs.GetParamUINT32(1)

	argv := (*[1 << 30]uint32)(cs.GetParamPointer(0))[:state.CmdLineNbArgs:state.CmdLineNbArgs]
	argvBuf := (*[1 << 30]byte)(cs.GetParamPointer(1))[:state.CmdLineArgsSize:state.CmdLineArgsSize]

	if state.wctx.Fctx.Trace {
		fmt.Printf("called args_get argv:%08x argvBuf:%08x p0:%x p1:%x\n", argv, argvBuf, p0, p1)
	}

	index := uint32(0)
	for argc, arg := range state.CmdLineArgs {
		argv[argc] = p1 + index

		copy(argvBuf[index:], []byte(arg))
		argvBuf[int(index)+len(arg)] = 0

		index = uint32(int(index) + len(arg) + 1)
	}

	return WASI_ESUCCESS, 0
}

// WASIFdPrestatGet WASIFdPrestatGet
func WASIFdPrestatGet(state *state, cs *CallSite) (uint32, int) {
	fd := cs.GetParamUINT32(0)
	buf := cs.GetParamPointer(1)

	if state.wctx.Fctx.Trace {
		fmt.Printf("called fd_prestat_get: fd %d buf %08x\n", fd, buf)
	}

	if fd < 3 || fd >= 5 {
		return WASI_EBADF, 0
	}

	*(*uint32)(buf) = WASI_PREOPENTYPE_DIR
	*(*uint32)(unsafe.Pointer(uintptr(buf) + uintptr(4))) = uint32(len(preopen[fd]))

	return WASI_ESUCCESS, 0
}

// WASIFdPrestatDirName WASIFdPrestatDirName
func WASIFdPrestatDirName(state *state, cs *CallSite) (uint32, int) {
	fd := cs.GetParamUINT32(0)
	buf := cs.GetParamPointer(1)
	length := cs.GetParamUINT32(2)

	if state.wctx.Fctx.Trace {
		fmt.Printf("called fd_prestat_dir_name: fd %d buf %x lenght %d\n", fd, buf, length)
	}

	if fd < 3 || fd >= 5 {
		return WASI_EBADF, 0
	}

	nameBuf := (*[1 << 30]byte)(buf)[:length:length]
	copy(nameBuf, []byte(preopen[fd]))

	return WASI_ESUCCESS, 0
}

// WASIFdFdStatGet WASIFdFdStatGet
func WASIFdFdStatGet(state *state, cs *CallSite) (uint32, int) {
	fd := cs.GetParamUINT32(0)
	fdStatAddr := cs.GetParamPointer(1)

	if state.wctx.Fctx.Trace {
		fmt.Printf("called fd_fdstat_get: fd %d fdStatAddr %x\n", fd, fdStatAddr)
	}

	setWasiStat(int(fd), fdStatAddr)

	return WASI_ESUCCESS, 0
}

// WASIPathFilestatGet WASIPathFilestatGet
func WASIPathFilestatGet(state *state, cs *CallSite) (uint32, int) {
	fd := cs.GetParamUINT32(0)
	lookupFlags := cs.GetParamUINT32(1)
	pathSlice := cs.GetParamByteBuffer(2, 3)
	fileStatAddr := (*C.__wasi_filestat_t)(cs.GetParamPointer(4))

	if state.wctx.Fctx.Trace {
		fmt.Printf("called path_filestat_get: fd %d lookupFlags %x path %s\n", fd, lookupFlags, string(pathSlice))
	}

	fileStatAddr.st_size = 30

	return WASI_ESUCCESS, 0
}

// WASIFdWrite WASIFdWrite
func WASIFdWrite(state *state, cs *CallSite) (uint32, int) {
	fd := cs.GetParamUINT32(0)
	wasiIovs := cs.GetParamPointer(1)
	iovsLen := cs.GetParamUINT32(2)
	nwritten := cs.GetParamPointer(3)

	if state.wctx.Fctx.Trace {
		fmt.Printf("called fd_write on fd %d: iovs: %x iovsLen: %d\n", fd, uintptr(wasiIovs), iovsLen)
	}

	writtenLength := uint32(0)
	for i := uint32(0); i < iovsLen; i++ {
		wasiIov := (*wasm3.WasiIoVec)(unsafe.Pointer(uintptr(wasiIovs) + unsafe.Sizeof(wasm3.WasiIoVec{})*uintptr(i)))
		addr := m3ApiOffsetToPtr(cs.mem, wasiIov.GetBuf())
		length := wasiIov.GetBufLen()

		buffer := (*[1 << 30]byte)(unsafe.Pointer(addr))[:length:length]

		if fd == 1 && state.wctx.Fctx.Trace {
			fmt.Printf(" [received iov:] '%s'\n", string(buffer))
		}

		virtualFile := state.OpenedVirtualFiles[int(fd)]
		l, err := virtualFile.Write(buffer)
		if err != nil {
			*(*uint32)(nwritten) = 0
			return WASI_EBADF, 0
		}
		writtenLength += uint32(l)
	}

	*(*uint32)(nwritten) = writtenLength

	// return value (was top of stack at the beginning of execution)
	return WASI_ESUCCESS, 0
}

// WASIFdRead WASIFdRead
func WASIFdRead(state *state, cs *CallSite) (uint32, int) {
	fd := cs.GetParamUINT32(0)
	wasiIovs := cs.GetParamPointer(1)
	iovsLen := cs.GetParamUINT32(2)
	nRead := cs.GetParamPointer(3)

	if state.wctx.Fctx.Trace {
		fmt.Printf("called fd_read on fd %d: iovs: %x iovsLen: %d\n", fd, uintptr(wasiIovs), iovsLen)
	}

	readLength := 0

	for i := uint32(0); i < iovsLen; i++ {
		wasiIov := (*wasm3.WasiIoVec)(unsafe.Pointer(uintptr(wasiIovs) + unsafe.Sizeof(wasm3.WasiIoVec{})*uintptr(i)))
		addr := m3ApiOffsetToPtr(cs.mem, wasiIov.GetBuf())
		length := int(wasiIov.GetBufLen())
		if length == 0 {
			continue
		}

		buffer := (*[1 << 30]byte)(unsafe.Pointer(addr))[:length:length]

		virtualFile := state.OpenedVirtualFiles[int(fd)]
		l := virtualFile.Read(buffer)
		readLength = readLength + l
	}

	*(*uint32)(nRead) = uint32(readLength)

	return WASI_ESUCCESS, 0
}

// WASIFdSeek WASIFdSeek
func WASIFdSeek(state *state, cs *CallSite) (uint32, int) {
	fd := cs.GetParamUINT32(0)
	offset := cs.GetParamUINT32(1)
	whence := cs.GetParamUINT32(2)
	resultAddr := cs.GetParamPointer(3)

	if state.wctx.Fctx.Trace {
		fmt.Printf("called fd_seek fd %d offset %d whence %d resultAddr %x\n", fd, offset, whence, resultAddr)
	}

	*(*uint32)(resultAddr) = offset

	return WASI_ESUCCESS, 0
}

func WASIFdClose(state *state, cs *CallSite) (uint32, int) {
	fd := cs.GetParamUINT32(0)

	virtualFile := state.OpenedVirtualFiles[int(fd)]
	virtualFile.Close()

	return WASI_ESUCCESS, 0
}

func WASIEnvironSizesGet(state *state, cs *CallSite) (uint32, int) {
	envCount := cs.GetParamUINT32Ptr(0)
	envBufSize := cs.GetParamUINT32Ptr(1)

	if state.wctx.Fctx.Trace {
		fmt.Printf("called environ_sizes_get with %x %x\n", envCount, envBufSize)
	}

	*envCount = 0
	*envBufSize = 0

	return WASI_ESUCCESS, 0
}

func WASIEnvironGet(state *state, cs *CallSite) (uint32, int) {
	if state.wctx.Fctx.Trace {
		fmt.Printf("called environ_get\n")
	}
	/*m3ApiReturnType  (uint32_t)
	    m3ApiGetArgMem   (u32*                 , env)
	    m3ApiGetArgMem   (char*                , env_buf)

	    if (runtime == NULL) { m3ApiReturn(__WASI_EINVAL); }
	    // TODO
		m3ApiReturn(__WASI_ESUCCESS);*/
	return WASI_ESUCCESS, 0
}
