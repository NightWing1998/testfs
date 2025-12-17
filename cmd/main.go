package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"path/filepath"

	mockfs "github.com/NightWing1998/testfs"
)

func playWithFS() {
	path := filepath.Join(os.TempDir(), "playwithfs")
	err := os.MkdirAll(path, 0755)
	if err != nil {
		log.Printf("mkdir all failed: %v", err)
	}
	defer func() {
		err := os.RemoveAll(path)
		if err != nil {
			log.Printf("remove all failed: %v", err)
		}
	}()
	fs := mockfs.New()
	fs.SetLogger(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: mockfs.LevelTrace})))
	if err := fs.Mount(context.Background(), path); err != nil {
		log.Fatalf("mount failed: %v", err)
	}
	defer fs.Close()

	// Observe filesystem activity.
	go func() {
		for ev := range fs.Events() {
			log.Printf("event=%s path=%s size=%d from=%s error=%v errno=%d", ev.Type, ev.Path, ev.Size, ev.OldPath, ev.Error, ev.Errno)
		}
	}()

	log.Printf("mounted filesystem at %s", path)

	{
		f, err := os.Create(filepath.Join(path, "test.txt"))
		if err != nil {
			log.Printf("create file failed: %v", err)
		}

		_, err = f.Write([]byte("hello, world"))
		if err != nil {
			log.Printf("write failed: %v", err)
		}

		err = f.Sync()
		if err != nil {
			log.Printf("sync failed: %v", err)
		}

		err = f.Close()
		if err != nil {
			log.Printf("close failed: %v", err)
		}

		log.Printf("file created and written to %s", filepath.Join(path, "test.txt"))
	}

	{
		f, err := os.Open(filepath.Join(path, "test.txt"))
		if err != nil {
			log.Printf("open file failed: %v", err)
		}

		buf := make([]byte, 1024)
		_, err = f.Read(buf)
		if err != nil {
			f.Close()
			log.Printf("read failed: %v", err)
		}
		f.Close()
	}

	{
		err := os.Mkdir(filepath.Join(path, "testdir"), 0755)
		if err != nil {
			log.Printf("mkdir failed: %v", err)
		}
		log.Printf("directory created %s", filepath.Join(path, "testdir"))

		info, err := os.Stat(filepath.Join(path, "testdir"))
		if err != nil {
			log.Printf("stat failed: %v", err)
		}
		log.Printf("stat for testdir: %+v", info)

		err = os.RemoveAll(filepath.Join(path, "testdir"))
		if err != nil {
			log.Printf("remove directory failed: %v", err)
		}

		log.Printf("directory removed %s", filepath.Join(path, "testdir"))
	}

	fs.InjectErrorOnEvent(mockfs.EventRead, mockfs.ErrIsDir)
	{
		f, err := os.Open(filepath.Join(path, "test.txt"))
		if err != nil {
			log.Printf("open file failed: %v", err)
		}

		buf := make([]byte, 1024)
		_, err = f.Read(buf)
		if err == nil {
			log.Printf("read expected to fail, but succeeded")
		}
		log.Printf("read failed with error: %v", err)

		info, err := f.Stat()
		if err != nil {
			log.Printf("stat failed: %v", err)
		}
		log.Printf("stat for test.txt: %+v", info)

		f.Close()

		os.Remove(filepath.Join(path, "test.txt"))

		log.Printf("file removed %s", filepath.Join(path, "test.txt"))
	}
}

func main() {
	playWithFS()
}
