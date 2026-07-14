package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	fmt.Printf("signal-demo started, pid=%d\n", os.Getpid())
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case now := <-ticker.C:
			fmt.Printf("working at %s\n", now.Format(time.RFC3339))
		case <-ctx.Done():
			fmt.Println("shutdown signal received; simulating cleanup")
			time.Sleep(2 * time.Second)
			fmt.Println("cleanup finished")
			return
		}
	}
}
