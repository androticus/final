package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql" // registers "mysql" driver
	_ "github.com/jackc/pgx/v5/stdlib" // registers "pgx" driver for database/sql
	_ "modernc.org/sqlite"             // registers "sqlite" driver (cznic / modernc, cgo-free)

	"github.com/androticus/final/finalembed"
	"github.com/androticus/final/finalgo"
	"github.com/androticus/final/finaltui"
)

func main() {
	headless := flag.Bool("headless", false, "skip tview; run stdin/stdout REPL after loads")
	userManifest := flag.String("manifest", "", "after platform load, compile this .fnlm or .fnl")
	flag.StringVar(userManifest, "m", "", "short for -manifest")
	stay := flag.Bool("i", false, "after a TTY script file, stay in stdin/stdout REPL")
	printVersion := flag.Bool("version", false, "print version string and exit")
	cpuProfile := flag.String("cpuprofile", "", "write a Go CPU profile to this path while the script runs")
	flag.Parse()
	if *printVersion {
		fmt.Println(finalgo.Version())
		return
	}

	spec, err := finaltui.ClassifyLaunch(finaltui.StdinIsTTY(), finaltui.LaunchFlags{
		Headless: *headless,
		Stay:     *stay,
		Manifest: *userManifest,
	}, flag.Args())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	sess, err := finalembed.NewPersistentSession()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer sess.Close()

	finaltui.ApplySysLaunch(os.Args, spec)

	if err := finaltui.Init(sess); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	headedTUI := spec.Mode == finaltui.LaunchInteractive && !spec.Headless
	if headedTUI {
		// Headed: SIGTERM exits; SIGINT → Break + cancel in-flight dialog. Quit is Ctrl-Q.
		go func() {
			for sig := range sigChan {
				if sig == syscall.SIGTERM {
					os.Exit(0)
				}
				sess.Break()
				finaltui.Tui_.Complete(finaltui.CompleteChoice(finaltui.BtnCancel))
			}
		}()
	} else {
		// Script, -i, or --headless: same soft/hard Ctrl-C policy as final-go.
		var doublePressTimer *time.Timer
		go func() {
			for {
				sig := <-sigChan
				if sig == syscall.SIGTERM {
					fmt.Printf("\nReceived SIGTERM - exiting\n")
					os.Exit(0)
				}
				if doublePressTimer != nil {
					doublePressTimer.Stop()
					doublePressTimer = nil
					fmt.Println("\nDouble Ctrl+C detected - forcing shutdown...")
					os.Exit(0)
					return
				}
				doublePressTimer = time.AfterFunc(600*time.Millisecond, func() {
					doublePressTimer = nil
					fmt.Printf("\nReceived %v - calling session.Break()\n", sig)
					sess.Break()
				})
			}
		}()
	}

	if *cpuProfile != "" {
		pf, err := os.Create(*cpuProfile)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := pprof.StartCPUProfile(pf); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer func() {
			pprof.StopCPUProfile()
			pf.Close()
		}()
	}

	if err := finaltui.RunLaunch(sess, spec, finaltui.HostOptions{}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
