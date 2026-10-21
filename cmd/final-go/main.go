package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql" // registers "mysql" driver
	_ "github.com/jackc/pgx/v5/stdlib" // registers "pgx" driver for database/sql
	_ "modernc.org/sqlite"             // registers "sqlite" driver (cznic / modernc, cgo-free)

	"github.com/androticus/final/finalgo"
)

func main() {
	useCore := flag.Bool("core", false, "use sealed core module instead of a per-run local module")
	restrict := flag.Int64("restrict", 0, "SecurityRestrictions disabled-set mask (uint32 themes; use named finalgo.SR_* or e.g. -restrict 0x200000 for SR_Threads)")
	printVersion := flag.Bool("version", false, "print version string and exit")
	flag.Parse()
	if *printVersion {
		fmt.Println(finalgo.Version())
		return
	}

	module := finalgo.REPLModuleLocal
	if *useCore {
		module = finalgo.REPLModuleCore
	}

	sess, err := finalgo.NewREPLSession(module, *restrict)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer sess.Close()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	var doublePressTimer *time.Timer
	go func() {
		for {
			sig := <-sigChan
			if sig == syscall.SIGTERM {
				fmt.Printf("\nReceived SIGTERM - exiting\n")
				os.Exit(0)
			}
			// SIGINT handling: 2 quick presses within 600ms: exit

			if doublePressTimer != nil {
				// Second press arrived before timer fired → double press
				doublePressTimer.Stop()
				doublePressTimer = nil
				fmt.Println("\nDouble Ctrl+C detected - forcing shutdown...")
				os.Exit(0)
				return
			}

			// First press: schedule the Break, but give user time to press again

			doublePressTimer = time.AfterFunc(600*time.Millisecond, func() {
				doublePressTimer = nil
				fmt.Printf("\nReceived %v - calling session.Break()\n", sig)
				sess.Break()
			})
		}
	}()

	if err := sess.RunREPL(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
