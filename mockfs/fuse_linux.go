//go:build linux
// +build linux

// Package mockfs provides the Linux FUSE-backed mock filesystem implementation.
package mockfs

/*
#cgo linux LDFLAGS: -lfuse3
#define FUSE_USE_VERSION 35
#include <fuse3/fuse.h>
#include <errno.h>
#include <stdlib.h>
#include <string.h>
#include <stdint.h>
#include <sys/statvfs.h>
#include <sys/types.h>

extern int go_getattr(char *path, struct stat *stbuf, struct fuse_file_info *fi, uint64_t handle);
extern int go_readdir(char *path, void *buf, fuse_fill_dir_t filler, off_t offset, struct fuse_file_info *fi, enum fuse_readdir_flags flags, uint64_t handle);
extern int go_open(char *path, struct fuse_file_info *fi, uint64_t handle);
extern int go_release(char *path, struct fuse_file_info *fi, uint64_t handle);
extern int go_read(char *path, char *buf, size_t size, off_t offset, struct fuse_file_info *fi, uint64_t handle);
extern int go_write(char *path, char *buf, size_t size, off_t offset, struct fuse_file_info *fi, uint64_t handle);
extern int go_create(char *path, mode_t mode, struct fuse_file_info *fi, uint64_t handle);
extern int go_unlink(char *path, uint64_t handle);
extern int go_mkdir(char *path, mode_t mode, uint64_t handle);
extern int go_rmdir(char *path, uint64_t handle);
extern int go_rename(char *oldpath, char *newpath, unsigned int flags, uint64_t handle);
extern int go_truncate(char *path, off_t size, struct fuse_file_info *fi, uint64_t handle);
extern int go_statfs(char *path, struct statvfs *st, uint64_t handle);

static int mockfs_getattr(const char *path, struct stat *stbuf, struct fuse_file_info *fi) {
    uint64_t h = *(uint64_t*)fuse_get_context()->private_data;
    return go_getattr((char *)path, stbuf, fi, h);
}

static int mockfs_readdir(const char *path, void *buf, fuse_fill_dir_t filler, off_t offset, struct fuse_file_info *fi, enum fuse_readdir_flags flags) {
    uint64_t h = *(uint64_t*)fuse_get_context()->private_data;
    return go_readdir((char *)path, buf, filler, offset, fi, flags, h);
}

static int mockfs_open(const char *path, struct fuse_file_info *fi) {
    uint64_t h = *(uint64_t*)fuse_get_context()->private_data;
    return go_open((char *)path, fi, h);
}

static int mockfs_release(const char *path, struct fuse_file_info *fi) {
    uint64_t h = *(uint64_t*)fuse_get_context()->private_data;
    return go_release((char *)path, fi, h);
}

static int mockfs_read(const char *path, char *buf, size_t size, off_t offset, struct fuse_file_info *fi) {
    uint64_t h = *(uint64_t*)fuse_get_context()->private_data;
    return go_read((char *)path, buf, size, offset, fi, h);
}

static int mockfs_write(const char *path, const char *buf, size_t size, off_t offset, struct fuse_file_info *fi) {
    uint64_t h = *(uint64_t*)fuse_get_context()->private_data;
    return go_write((char *)path, (char *)buf, size, offset, fi, h);
}

static int mockfs_create(const char *path, mode_t mode, struct fuse_file_info *fi) {
    uint64_t h = *(uint64_t*)fuse_get_context()->private_data;
    return go_create((char *)path, mode, fi, h);
}

static int mockfs_unlink(const char *path) {
    uint64_t h = *(uint64_t*)fuse_get_context()->private_data;
    return go_unlink((char *)path, h);
}

static int mockfs_mkdir(const char *path, mode_t mode) {
    uint64_t h = *(uint64_t*)fuse_get_context()->private_data;
    return go_mkdir((char *)path, mode, h);
}

static int mockfs_rmdir(const char *path) {
    uint64_t h = *(uint64_t*)fuse_get_context()->private_data;
    return go_rmdir((char *)path, h);
}

static int mockfs_rename(const char *oldpath, const char *newpath, unsigned int flags) {
    uint64_t h = *(uint64_t*)fuse_get_context()->private_data;
    return go_rename((char *)oldpath, (char *)newpath, flags, h);
}

static int mockfs_truncate(const char *path, off_t size, struct fuse_file_info *fi) {
    uint64_t h = *(uint64_t*)fuse_get_context()->private_data;
    return go_truncate((char *)path, size, fi, h);
}

static int mockfs_statfs(const char *path, struct statvfs *st) {
    uint64_t h = *(uint64_t*)fuse_get_context()->private_data;
    return go_statfs((char *)path, st, h);
}

static int mockfs_call_filler(fuse_fill_dir_t filler, void *buf, const char *name, const struct stat *stbuf, off_t off, enum fuse_fill_dir_flags flags) {
    return filler(buf, name, stbuf, off, flags);
}

struct mockfs_handle {
    struct fuse *fuse;
    char *mountpoint;
    uint64_t *user_handle;
};

static struct fuse_operations mockfs_ops = {
    .getattr = mockfs_getattr,
    .readdir = mockfs_readdir,
    .open = mockfs_open,
    .release = mockfs_release,
    .read = mockfs_read,
    .write = mockfs_write,
    .create = mockfs_create,
    .unlink = mockfs_unlink,
    .mkdir = mockfs_mkdir,
    .rmdir = mockfs_rmdir,
    .rename = mockfs_rename,
    .truncate = mockfs_truncate,
    .statfs = mockfs_statfs,
};

static void cleanup_handle(struct mockfs_handle *h) {
    if (!h) return;
    if (h->mountpoint) free(h->mountpoint);
    if (h->user_handle) free(h->user_handle);
    free(h);
}

static struct mockfs_handle* mockfs_new_handle(const char *mountpoint, uint64_t handle, char **err_out) {
    struct mockfs_handle *h = calloc(1, sizeof(struct mockfs_handle));
    if (!h) {
        if (err_out) *err_out = strdup("allocation failed");
        return NULL;
    }
    h->mountpoint = strdup(mountpoint);
    h->user_handle = malloc(sizeof(uint64_t));
    if (!h->mountpoint || !h->user_handle) {
        if (err_out) *err_out = strdup("allocation failed");
        cleanup_handle(h);
        return NULL;
    }
    *(h->user_handle) = handle;

    struct fuse_args args = FUSE_ARGS_INIT(0, NULL);
    // FUSE requires argv[0] to be present; supply a placeholder name.
    fuse_opt_add_arg(&args, "mockfs");
    h->fuse = fuse_new(&args, &mockfs_ops, sizeof(mockfs_ops), h->user_handle);
    if (!h->fuse) {
        if (err_out) *err_out = strdup("fuse_new failed");
        fuse_opt_free_args(&args);
        cleanup_handle(h);
        return NULL;
    }
    if (fuse_mount(h->fuse, mountpoint) != 0) {
        if (err_out) *err_out = strdup("fuse_mount failed");
        fuse_destroy(h->fuse);
        fuse_opt_free_args(&args);
        cleanup_handle(h);
        return NULL;
    }
    fuse_opt_free_args(&args);
    return h;
}

static int mockfs_loop(struct mockfs_handle *h) {
    if (!h) return -1;
    return fuse_loop(h->fuse);
}

static void mockfs_exit(struct mockfs_handle *h) {
    if (h && h->fuse) {
        fuse_exit(h->fuse);
    }
}

static void mockfs_destroy(struct mockfs_handle *h) {
    if (!h) return;
    if (h->fuse) {
        fuse_unmount(h->fuse);
        fuse_destroy(h->fuse);
    }
    cleanup_handle(h);
}
*/
import "C"

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"runtime/cgo"
	"strings"
	"sync"
	"time"
	"unsafe"
)

type fuseMount struct {
	handle     cgo.Handle
	native     *C.struct_mockfs_handle
	mountpoint string
	loopCh     chan struct{}
	closeMu    sync.Mutex
	closed     bool
}

// Mount mounts the mock filesystem at the given mountpoint and starts serving
// requests. The call returns once the mount is established. The provided
// context may be used to request unmount; canceling the context will trigger
// a clean shutdown.
func (fs *MockFS) Mount(ctx context.Context, mountpoint string) error {
	go fs.logLoop()

	fs.mountMu.Lock()
	defer fs.mountMu.Unlock()

	if _, ok := fs.platformData.(*fuseMount); ok {
		fs.logError("mount requested but already mounted", slog.String("mountpoint", mountpoint))
		return fmt.Errorf("mockfs is already mounted")
	}

	fs.logInfo("mounting mockfs", slog.String("mountpoint", mountpoint))

	h := cgo.NewHandle(fs)
	cMount := C.CString(mountpoint)
	defer C.free(unsafe.Pointer(cMount))

	var errStr *C.char
	native := C.mockfs_new_handle(cMount, C.uint64_t(uintptr(h)), &errStr)
	if errStr != nil {
		defer C.free(unsafe.Pointer(errStr))
	}
	if native == nil {
		h.Delete()
		msg := "failed to create fuse instance"
		if errStr != nil {
			msg = fmt.Sprintf("%s: %s", msg, C.GoString(errStr))
		}
		fs.logError(msg, slog.String("mountpoint", mountpoint))
		return errors.New(msg)
	}

	m := &fuseMount{
		handle:     h,
		native:     native,
		mountpoint: mountpoint,
		loopCh:     make(chan struct{}),
	}
	fs.platformData = m
	fs.logInfo("mockfs mounted", slog.String("mountpoint", mountpoint))

	go func() {
		C.mockfs_loop(native)
		fs.logInfo("fuse loop exited", slog.String("mountpoint", mountpoint))
		close(m.loopCh)
	}()

	if ctx != nil {
		go func() {
			<-ctx.Done()
			_ = fs.Close()
		}()
	}
	return nil
}

// Close unmounts the filesystem and releases resources. It is safe to call
// multiple times.
func (fs *MockFS) Close() error {
	fs.mountMu.Lock()
	defer fs.mountMu.Unlock()

	fm, ok := fs.platformData.(*fuseMount)
	if !ok {
		return nil
	}
	fm.closeMu.Lock()
	if fm.closed {
		fm.closeMu.Unlock()
		return nil
	}
	fm.closed = true
	fm.closeMu.Unlock()

	fs.logInfo("closing mockfs", slog.String("mountpoint", fm.mountpoint))

	// Signal FUSE to exit
	C.mockfs_exit(fm.native)

	// Wait for the loop to exit with a timeout to prevent indefinite blocking
	select {
	case <-fm.loopCh:
		// Loop exited normally
		fs.logInfo("mockfs closed", slog.String("mountpoint", fm.mountpoint))
	case <-time.After(5 * time.Second):
		// Loop didn't exit in time - this could indicate:
		// 1. mockfs_exit didn't properly signal the loop
		// 2. The loop is blocked on an operation
		// 3. There's a race condition in the C code
		// We proceed anyway to avoid deadlock, but log this condition
		// Note: This may leak resources if the loop is truly stuck
		fs.logError("fuse loop did not exit before timeout; forcing shutdown", slog.String("mountpoint", fm.mountpoint))
	}

	C.mockfs_destroy(fm.native)
	fm.handle.Delete()
	close(fs.events)
	fs.platformData = nil
	close(fs.logCh)
	return nil
}

func handleToFS(h C.uint64_t) *MockFS {
	return cgo.Handle(uintptr(h)).Value().(*MockFS)
}

//export go_getattr
func go_getattr(path *C.char, stbuf *C.struct_stat, fi *C.struct_fuse_file_info, handle C.uint64_t) C.int {
	_ = fi
	fs := handleToFS(handle)
	goPath := C.GoString(path)
	fs.logTrace("go_getattr start", slog.String("path", goPath))
	node, err := fs.lookup(goPath)
	if err != nil {
		fs.logDebug("go_getattr failed", slog.String("path", goPath), slog.Any("error", err))
		return -errnoFrom(err)
	}
	fillStat(stbuf, node)
	fs.logTrace("go_getattr success", slog.String("path", goPath))
	return 0
}

//export go_readdir
func go_readdir(path *C.char, buf unsafe.Pointer, filler C.fuse_fill_dir_t, offset C.off_t, fi *C.struct_fuse_file_info, flags C.enum_fuse_readdir_flags, handle C.uint64_t) C.int {
	_ = fi
	_ = offset
	_ = flags
	fs := handleToFS(handle)
	goPath := C.GoString(path)
	fs.logTrace("go_readdir start", slog.String("path", goPath))
	entries, err := fs.list(goPath)
	if err != nil {
		fs.logDebug("go_readdir failed", slog.String("path", goPath), slog.Any("error", err))
		return -errnoFrom(err)
	}

	for _, name := range []string{".", ".."} {
		cname := C.CString(name)
		C.mockfs_call_filler(filler, buf, cname, nil, 0, 0)
		C.free(unsafe.Pointer(cname))
	}

	for _, n := range entries {
		cname := C.CString(n.name)
		C.mockfs_call_filler(filler, buf, cname, nil, 0, 0)
		C.free(unsafe.Pointer(cname))
	}
	fs.logTrace("go_readdir success", slog.String("path", goPath), slog.Int("count", len(entries)))
	return 0
}

//export go_open
func go_open(path *C.char, fi *C.struct_fuse_file_info, handle C.uint64_t) C.int {
	fs := handleToFS(handle)
	goPath := C.GoString(path)
	fs.logTrace("go_open start", slog.String("path", goPath))
	fh, _, err := fs.open(goPath)
	if err != nil {
		errno := errnoFrom(err)
		fs.emit(Event{Type: EventOpen, Path: goPath, Error: err, Errno: int(errno)})
		fs.logDebug("go_open failed", slog.String("path", goPath), slog.Any("error", err))
		return -errno
	}
	fi.fh = C.uint64_t(fh)
	fs.emit(Event{Type: EventOpen, Path: goPath})
	fs.logTrace("go_open success", slog.String("path", goPath), slog.Uint64("fh", uint64(fh)))
	return 0
}

//export go_release
func go_release(path *C.char, fi *C.struct_fuse_file_info, handle C.uint64_t) C.int {
	_ = path
	fs := handleToFS(handle)
	fs.logTrace("go_release start", slog.Uint64("fh", uint64(fi.fh)))
	fs.releaseHandle(uint64(fi.fh))
	fs.logTrace("go_release success", slog.Uint64("fh", uint64(fi.fh)))
	return 0
}

//export go_read
func go_read(path *C.char, buf *C.char, size C.size_t, offset C.off_t, fi *C.struct_fuse_file_info, handle C.uint64_t) C.int {
	fs := handleToFS(handle)
	goPath := C.GoString(path)
	fs.logTrace("go_read start", slog.String("path", goPath), slog.Int("size", int(size)), slog.Int64("offset", int64(offset)), slog.Uint64("fh", uint64(fi.fh)))
	var n *node
	if fhNode, ok := fs.handleNode(uint64(fi.fh)); ok {
		n = fhNode
	} else {
		var err error
		n, err = fs.lookup(goPath)
		if err != nil {
			errno := errnoFrom(err)
			fs.emit(Event{Type: EventRead, Path: goPath, Error: err, Errno: int(errno)})
			fs.logDebug("go_read lookup failed", slog.String("path", goPath), slog.Any("error", err))
			return -errno
		}
	}
	data, err := fs.readFile(n, int64(offset), int(size))
	if err != nil {
		errno := errnoFrom(err)
		fs.emit(Event{Type: EventRead, Path: goPath, Error: err, Errno: int(errno)})
		fs.logDebug("go_read failed", slog.String("path", goPath), slog.Any("error", err))
		return -errno
	}
	if len(data) > 0 {
		C.memcpy(unsafe.Pointer(buf), unsafe.Pointer(&data[0]), C.size_t(len(data)))
	}
	fs.emit(Event{Type: EventRead, Path: goPath, Size: int64(len(data))})
	fs.logTrace("go_read success", slog.String("path", goPath), slog.Int("read", len(data)), slog.Uint64("fh", uint64(fi.fh)))
	return C.int(len(data))
}

//export go_write
func go_write(path *C.char, cbuf *C.char, size C.size_t, offset C.off_t, fi *C.struct_fuse_file_info, handle C.uint64_t) C.int {
	fs := handleToFS(handle)
	goPath := C.GoString(path)
	fs.logTrace("go_write start", slog.String("path", goPath), slog.Int("size", int(size)), slog.Int64("offset", int64(offset)), slog.Uint64("fh", uint64(fi.fh)))
	var n *node
	if fhNode, ok := fs.handleNode(uint64(fi.fh)); ok {
		n = fhNode
	} else {
		var err error
		n, err = fs.lookup(goPath)
		if err != nil {
			errno := errnoFrom(err)
			fs.emit(Event{Type: EventWrite, Path: goPath, Error: err, Errno: int(errno)})
			fs.logDebug("go_write lookup failed", slog.String("path", goPath), slog.Any("error", err))
			return -errno
		}
	}
	data := C.GoBytes(unsafe.Pointer(cbuf), C.int(size))
	written, err := fs.writeFile(n, data, int64(offset))
	if err != nil {
		errno := errnoFrom(err)
		fs.emit(Event{Type: EventWrite, Path: goPath, Error: err, Errno: int(errno)})
		fs.logDebug("go_write failed", slog.String("path", goPath), slog.Any("error", err))
		return -errno
	}
	fs.emit(Event{Type: EventWrite, Path: goPath, Size: int64(written)})
	fs.logTrace("go_write success", slog.String("path", goPath), slog.Int("written", written), slog.Uint64("fh", uint64(fi.fh)))
	return C.int(written)
}

//export go_create
func go_create(path *C.char, mode C.mode_t, fi *C.struct_fuse_file_info, handle C.uint64_t) C.int {
	fs := handleToFS(handle)
	goPath := C.GoString(path)
	fs.logTrace("go_create start", slog.String("path", goPath), slog.Uint64("mode", uint64(mode)))
	n, err := fs.createFile(goPath, os.FileMode(mode)&0o777)
	if err != nil {
		errno := errnoFrom(err)
		fs.emit(Event{Type: EventCreate, Path: goPath, Error: err, Errno: int(errno)})
		fs.logDebug("go_create failed", slog.String("path", goPath), slog.Any("error", err))
		return -errno
	}
	fh, _, err := fs.open(goPath)
	if err != nil {
		errno := errnoFrom(err)
		fs.emit(Event{Type: EventCreate, Path: goPath, Error: err, Errno: int(errno)})
		fs.logDebug("go_create open failed", slog.String("path", goPath), slog.Any("error", err))
		return -errno
	}
	fi.fh = C.uint64_t(fh)
	fs.emit(Event{Type: EventCreate, Path: fullPath(n)})
	fs.logTrace("go_create success", slog.String("path", goPath), slog.Uint64("fh", uint64(fh)))
	return 0
}

//export go_unlink
func go_unlink(path *C.char, handle C.uint64_t) C.int {
	fs := handleToFS(handle)
	goPath := C.GoString(path)
	fs.logTrace("go_unlink start", slog.String("path", goPath))
	if err := fs.unlink(goPath); err != nil {
		errno := errnoFrom(err)
		fs.emit(Event{Type: EventRemove, Path: goPath, Error: err, Errno: int(errno)})
		fs.logDebug("go_unlink failed", slog.String("path", goPath), slog.Any("error", err))
		return -errno
	}
	fs.emit(Event{Type: EventRemove, Path: goPath})
	fs.logTrace("go_unlink success", slog.String("path", goPath))
	return 0
}

//export go_mkdir
func go_mkdir(path *C.char, mode C.mode_t, handle C.uint64_t) C.int {
	fs := handleToFS(handle)
	goPath := C.GoString(path)
	fs.logTrace("go_mkdir start", slog.String("path", goPath), slog.Uint64("mode", uint64(mode)))
	if err := fs.mkdir(goPath, os.FileMode(mode)&0o777); err != nil {
		errno := errnoFrom(err)
		fs.emit(Event{Type: EventMkdir, Path: goPath, Error: err, Errno: int(errno)})
		fs.logDebug("go_mkdir failed", slog.String("path", goPath), slog.Any("error", err))
		return -errno
	}
	fs.emit(Event{Type: EventMkdir, Path: goPath})
	fs.logTrace("go_mkdir success", slog.String("path", goPath))
	return 0
}

//export go_rmdir
func go_rmdir(path *C.char, handle C.uint64_t) C.int {
	fs := handleToFS(handle)
	goPath := C.GoString(path)
	fs.logTrace("go_rmdir start", slog.String("path", goPath))
	if err := fs.rmdir(goPath); err != nil {
		errno := errnoFrom(err)
		fs.emit(Event{Type: EventRmdir, Path: goPath, Error: err, Errno: int(errno)})
		fs.logDebug("go_rmdir failed", slog.String("path", goPath), slog.Any("error", err))
		return -errno
	}
	fs.emit(Event{Type: EventRmdir, Path: goPath})
	fs.logTrace("go_rmdir success", slog.String("path", goPath))
	return 0
}

//export go_rename
func go_rename(oldpath *C.char, newpath *C.char, flags C.uint, handle C.uint64_t) C.int {
	_ = flags
	fs := handleToFS(handle)
	oldGo := C.GoString(oldpath)
	newGo := C.GoString(newpath)
	fs.logTrace("go_rename start", slog.String("old_path", oldGo), slog.String("new_path", newGo))
	if err := fs.rename(oldGo, newGo); err != nil {
		errno := errnoFrom(err)
		fs.emit(Event{Type: EventRename, Path: newGo, OldPath: oldGo, Error: err, Errno: int(errno)})
		fs.logDebug("go_rename failed", slog.String("old_path", oldGo), slog.String("new_path", newGo), slog.Any("error", err))
		return -errno
	}
	fs.emit(Event{Type: EventRename, Path: newGo, OldPath: oldGo})
	fs.logTrace("go_rename success", slog.String("old_path", oldGo), slog.String("new_path", newGo))
	return 0
}

//export go_truncate
func go_truncate(path *C.char, size C.off_t, fi *C.struct_fuse_file_info, handle C.uint64_t) C.int {
	_ = fi
	fs := handleToFS(handle)
	goPath := C.GoString(path)
	fs.logTrace("go_truncate start", slog.String("path", goPath), slog.Int64("size", int64(size)))
	if err := fs.truncate(goPath, int64(size)); err != nil {
		errno := errnoFrom(err)
		fs.emit(Event{Type: EventTruncate, Path: goPath, Size: int64(size), Error: err, Errno: int(errno)})
		fs.logDebug("go_truncate failed", slog.String("path", goPath), slog.Any("error", err))
		return -errno
	}
	fs.emit(Event{Type: EventTruncate, Path: goPath, Size: int64(size)})
	fs.logTrace("go_truncate success", slog.String("path", goPath), slog.Int64("size", int64(size)))
	return 0
}

//export go_statfs
func go_statfs(path *C.char, st *C.struct_statvfs, handle C.uint64_t) C.int {
	fs := handleToFS(handle)
	goPath := C.GoString(path)
	fs.logTrace("go_statfs start", slog.String("path", goPath))

	// Default fallback values if node is not found
	bsize := C.ulong(4096)
	frsize := C.ulong(4096)
	var blocks, bfree, bavail, files, ffree C.ulong = 0, 0, 0, 0, 0

	n, err := fs.lookup(goPath)
	if err != nil {
		return -errnoFrom(err)
	}
	if n != nil {
		// total size in "blocks"
		files = C.ulong(1)
		blocks = C.ulong(len(n.data)+int(bsize)-1) / bsize
		bfree = blocks // all blocks "free" (mock)
		bavail = blocks
		ffree = files
	} else {
		// fallback to 1 block/file
		files = 1
		blocks = 1
		bfree = 1
		bavail = 1
		ffree = 1
	}

	// Zero out struct
	var zero C.struct_statvfs
	*st = zero

	st.f_bsize = bsize
	st.f_frsize = frsize
	st.f_blocks = blocks
	st.f_bfree = bfree
	st.f_bavail = bavail
	st.f_files = files
	st.f_ffree = ffree
	st.f_namemax = 255

	fs.logTrace("go_statfs success", slog.String("path", goPath), slog.Int64("blocks", int64(blocks)), slog.Int64("files", int64(files)))
	return 0
}

func fillStat(st *C.struct_stat, n *node) {
	C.memset(unsafe.Pointer(st), 0, C.size_t(unsafe.Sizeof(*st)))
	if n.isDir() {
		st.st_mode = C.S_IFDIR | C.mode_t(n.mode&0o777)
		st.st_nlink = 2
	} else {
		st.st_mode = C.S_IFREG | C.mode_t(n.mode&0o777)
		st.st_nlink = 1
		st.st_size = C.off_t(len(n.data))
	}
	ts := C.struct_timespec{
		tv_sec:  C.time_t(n.mtime.Unix()),
		tv_nsec: C.long(n.mtime.Nanosecond()),
	}
	st.st_atim = ts
	st.st_mtim = ts
	st.st_ctim = ts
	st.st_uid = C.uid_t(os.Getuid())
	st.st_gid = C.gid_t(os.Getgid())
}

func errnoFrom(err error) C.int {
	switch {
	case errors.Is(err, os.ErrNotExist):
		return C.ENOENT
	case errors.Is(err, ErrNotDir):
		return C.ENOTDIR
	case errors.Is(err, ErrExists):
		return C.EEXIST
	case errors.Is(err, ErrIsDir):
		return C.EISDIR
	case errors.Is(err, ErrNotEmpty):
		return C.ENOTEMPTY
	case errors.Is(err, ErrInvalid):
		return C.EINVAL
	default:
		return C.EIO
	}
}

func fullPath(n *node) string {
	if n == nil {
		return "/"
	}
	var parts []string
	cur := n
	for cur != nil && cur.name != "" {
		parts = append([]string{cur.name}, parts...)
		cur = cur.parent
	}
	return "/" + strings.Join(parts, "/")
}
