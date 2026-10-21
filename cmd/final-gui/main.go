package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"

	"github.com/androticus/final/finalembed"
	"github.com/androticus/final/finalgui"
	"github.com/androticus/final/internal/fyneui"
)

func init() {
	runtime.LockOSThread()
}

func main() {
	printVersion := flag.Bool("version", false, "print version string and exit")
	flag.Parse()
	if *printVersion {
		fmt.Println(finalgui.Version())
		return
	}

	sess, err := finalembed.NewPersistentSession()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer sess.Close()

	if err := finalgui.Init(sess); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	win := finalgui.BuildChrome()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for sig := range sigChan {
			if sig == syscall.SIGTERM {
				os.Exit(0)
			}
			sess.Break()
		}
	}()

	loopErr := make(chan error, 1)
	go func() {
		loopErr <- finalgui.Run(sess)
	}()

	win.Show()
	fyneui.TheApp.Run()
	finalgui.Stop()
	if err := <-loopErr; err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
