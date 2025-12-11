# testfs: minimal FUSE-backed mock filesystem

An in-memory filesystem intended for testing filesystem-aware code in golang.
It mounts via the Linux FUSE kernel module and emits a small set of events
that mirror common file operations. Only standard library Go code plus `libfuse3`
is used; there are no additional third-party Go dependencies. Non-Linux builds
compile to stubs that return an unsupported error.

## Requirements

- Linux with FUSE 3 kernel support
- `libfuse3` development headers (`apt install libfuse3-dev` or equivalent)
- CGO enabled (`CGO_ENABLED=1`)

## Installing

```
go get github.com/NightWing1998/testfs
```

## Usage

```go
package main

import (
	"context"
	"log"
	"time"

	mockfs "github.com/NightWing1998/testfs"
)

func main() {
	fs := mockfs.New()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Mount the in-memory filesystem at /tmp/mockfs (must exist).
	if err := fs.Mount(ctx, "/tmp/mockfs"); err != nil {
		log.Fatalf("mount failed: %v", err)
	}
	defer fs.Close()

	// Observe filesystem activity.
	go func() {
		for ev := range fs.Events() {
			log.Printf("event=%s path=%s size=%d from=%s", ev.Type, ev.Path, ev.Size, ev.OldPath)
		}
	}()

	// Keep running; interact with the mountpoint using normal shell commands.
	time.Sleep(10 * time.Second)
}
```

Supported operations include `getattr`, `readdir`, `create`, `open`, `read`,
`write`, `truncate`, `unlink`, `mkdir`, `rmdir`, `rename`, and `statfs`. Events
are buffered (depth 64) and dropped if the buffer is full. Call `Close` (or
cancel the context passed to `Mount`) to unmount and clean up.

## Testing manually

1. Create a mount directory: `mkdir -p /tmp/mockfs`.
2. Run the sample above and in another terminal perform `touch`, `echo >>`,
   `cat`, `rm`, `mv`, `mkdir`, and `rmdir` under `/tmp/mockfs`.
3. Watch the emitted events to confirm your code responds as expected.

## Inject Errors

The purpose of testfs is to create a mock filesystem where you can inject
errors for specific filesystem events and verify your application's error
handling logic.

To inject errors, use the `InjectErrorOnEvent` method on the `MockFS` instance.
Here is an example that injects a read error:

```go
fs := testfs.New()

// Inject an error when a "read" event occurs
fs.InjectErrorOnEvent(testfs.EventRead, testfs.ErrIsDir)

// Now any read operation will emit this as an error
```

Supported event types you can use with `InjectErrorOnEvent` are:

- `testfs.EventCreate`   (file create)
- `testfs.EventWrite`    (file write)
- `testfs.EventRemove`   (file remove)
- `testfs.EventRead`     (file read)
- `testfs.EventRename`   (file rename)
- `testfs.EventMkdir`    (directory create)
- `testfs.EventRmdir`    (directory remove)
- `testfs.EventOpen`     (open file/directory)
- `testfs.EventTruncate` (truncate file)

Supported error types are:

- `testfs.ErrNotDir`   (not a directory error)
- `testfs.ErrExists`   (file or directory already exists)
- `testfs.ErrIsDir`    (is a directory error)
- `testfs.ErrNotEmpty` (directory not empty)
- `testfs.ErrNotFile`  (not a regular file)
- `testfs.ErrInvalid`  (invalid argument)

You can inject and clear errors at any time during the test to simulate various failure conditions.

## NOTE

> **WARNING:**
>
> If there are any file handles or open pointers still accessing the mount point or its children,
> the mount point may remain busy and cannot be unmounted cleanly by fuse. Future use for the same mount
> point can give errors like -
> ```log
> Transport endpoint is not connected
> ```
> In such cases, you must manually
> clear the mount using:
>
> ```sh
> umount <path>
> ```
