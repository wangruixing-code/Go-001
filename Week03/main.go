package main

import (
	"context"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx := context.Background()
	g, cancel := errgroup.WithContext(ctx)
	g.Go(func() error {
		return recvSignal(cancel)
	})
	g.Go(func() error {
		return runServer(cancel, ":8080")
	})
	if err := g.Wait(); err != nil {
		log.Println("error group: ", err.Error())
	}

	log.Println("exit.")
}

func recvSignal(ctx context.Context) error {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	select {
	case <-signalChan:
		return errors.New("exit -> close signal")
	case <-ctx.Done():
		return errors.New("exit -> ctx done")
	}
}

func runServer(ctx context.Context, addr string) error {
	srv := &http.Server{Addr: addr, Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "hello\n")
	})}
	go func() {
		select {
		case <-ctx.Done():
			shutdownServer(srv, ctx)
		}
	}()
	log.Println("running at ", addr)
	return srv.ListenAndServe()
}

func shutdownServer(server *http.Server, ctx context.Context) {
	_ = server.Shutdown(ctx)
}
