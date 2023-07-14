package main

import (
	"context"
	"github.com/Gentleelephant/EnhancementWorkload/cmd/apiserver/app/options"
	"k8s.io/klog/v2"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	serverRunOptions := options.NewServerRunOptions()

	err := Run(context.Background(), serverRunOptions)
	if err != nil {
		klog.Fatal(err)
	}
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	<-sig
}

func Run(ctx context.Context, s *options.ServerRunOptions) error {

	server, err := s.NewApiServer(ctx.Done())
	if err != nil {
		return err
	}
	err = server.PrepareRun(ctx.Done())
	if err != nil {
		return err
	}
	err = server.Run(ctx)
	if err != nil {
		klog.Error(err)
	}

	return nil

}
