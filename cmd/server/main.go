package main

import (
	"context"
	"fmt"
	"net"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	server "github.com/georgemac/grpc-fail"
	"github.com/influxdata/influxdb/v2/kit/cli"
	influxlogger "github.com/influxdata/influxdb/v2/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
)

var argv struct {
	LogLevel    string
	GRPCAddress string
}

func main() {
	prog := &cli.Program{
		Run:  run,
		Name: "inspectd",
		Opts: []cli.Opt{
			{
				DestP:   &argv.LogLevel,
				Flag:    "log-level",
				Default: "info",
				Desc:    "supported log levels are debug, info, and error",
			},
			{
				DestP:   &argv.GRPCAddress,
				Flag:    "grpc-address",
				Default: ":8002",
				Desc:    "bind address for the rest http api",
			},
		},
	}
	cmd := cli.NewCommand(prog)
	if err := cmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "inspectd error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	log, err := newLogger(argv.LogLevel)
	if err != nil {
		return err
	}
	log.Info("Running server")

	var (
		terminated  = make(chan os.Signal, 1)
		ctx, cancel = context.WithCancel(context.Background())
	)
	signal.Notify(terminated, syscall.SIGINT, syscall.SIGTERM)

	defer func() {
		signal.Stop(terminated)
		close(terminated)
		cancel()
	}()

	go func() {
		sig := <-terminated
		log.Info("Terminated", zap.Stringer("signal", sig))
		cancel()
	}()

	s := server.NewServer(5 * time.Second)

	grpcServer, grpcServerError := grpc.NewServer(), make(chan error)
	{
		server.RegisterSlowServiceServer(grpcServer, s)
		listener, err := net.Listen("tcp", argv.GRPCAddress)
		if err != nil {
			return err
		}
		go func() {
			if err := grpcServer.Serve(listener); err != grpc.ErrServerStopped {
				grpcServerError <- err
			}
		}()
	}

	select {
	case err = <-grpcServerError:
		log.Error("gRPC server error", zap.Error(err))
	case <-ctx.Done():
		shutdownGRPC(context.Background(), log, grpcServer)
	}
	if err != nil {
		return err
	}

	log.Info("Shutdown gracefully")
	return nil
}

func newLogger(logLevel string) (*zap.Logger, error) {
	var lvl zapcore.Level
	err := lvl.Set(logLevel)
	if err != nil {
		return nil, fmt.Errorf("unknown log level; supported levels are debug, info, and error")
	}
	config := influxlogger.NewConfig()
	config.Level = lvl
	log, err := config.New(os.Stdout)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %v", err)
	}

	return log, nil
}

// Attempt to stop the gRPC server gracefully before forcing it to stop after a timeout.
func shutdownGRPC(ctx context.Context, log *zap.Logger, grpcServer *grpc.Server) {
	log.Info("Shutting down gRPC server...")
	timeout, cancelTimeout := context.WithTimeout(ctx, 10*time.Second)
	stopped := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		stopped <- struct{}{}
	}()

	select {
	case <-stopped:
	case <-timeout.Done():
		log.Error("Failed to shutdown gRPC server gracefully")
		grpcServer.Stop()
	}

	cancelTimeout()
	log.Info("gRPC server shutdown")
}
