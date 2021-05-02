package main

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"net/http"
	"os"
	"os/signal"
)

func serverApp(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("pong"))
	})

	return server("127.0.0.1:8080", mux, ctx)
}

func server(addr string, handler http.Handler, ctx context.Context) error {
	srv := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	go func() {
		<-ctx.Done()
		srv.Shutdown(context.Background())
	}()

	return srv.ListenAndServe()
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	group, errCtx := errgroup.WithContext(ctx)

	group.Go(func() error {
		defer cancel()
		return serverApp(errCtx)
	})

	group.Go(func() error {
		defer cancel()
		return server("127.0.0.1:8081", http.DefaultServeMux, errCtx)
	})

	group.Go(func() error {
		quit := make(chan os.Signal)
		signal.Notify(quit, os.Interrupt)
		select {
		case <-quit:
			cancel()
		case <-errCtx.Done():
		}
		return nil
	})

	if err := group.Wait(); err != nil {
		fmt.Println("Get errors: ", err)
	} else {
		fmt.Println("end")
	}
}
