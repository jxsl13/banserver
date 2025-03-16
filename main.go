package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/jxsl13/banserver/config"
	"github.com/jxsl13/banserver/model"
	"github.com/jxsl13/cli-config-boilerplate/cliconfig"
	"github.com/spf13/cobra"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cmd := NewRootCmd(ctx)
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func NewRootCmd(ctx context.Context) *cobra.Command {
	root := RootContext{
		ctx: ctx,
		cfg: config.New(),
	}

	cmd := cobra.Command{
		Use:   filepath.Base(os.Args[0]),
		Short: "twlog is a utility for analyzing Teeworlds server logs",
	}

	cmd.PreRunE = root.PreRunE(&cmd)
	cmd.RunE = root.RunE
	cmd.AddCommand(NewCompletionCommand(&cmd))

	return &cmd
}

type RootContext struct {
	ctx context.Context
	cfg *config.Config
}

func (cli *RootContext) PreRunE(cmd *cobra.Command) func(*cobra.Command, []string) error {
	cfgParser := cliconfig.RegisterFlags(cli.cfg, false, cmd)
	return func(cmd *cobra.Command, args []string) error {
		log.SetOutput(cmd.ErrOrStderr()) // redirect log output to stderr
		return cfgParser()
	}
}

func (cli *RootContext) PostRunE(*cobra.Command, []string) error {
	log.Println("banserver shut down successfully")
	return nil
}

func (cli *RootContext) RunE(*cobra.Command, []string) (err error) {
	log.Println("starting banserver...")

	broker := model.NewBroker(
		cli.cfg.Propagate,
		cli.cfg.PermaBanDuration,
		cli.cfg.PermaBanReason,
		cli.cfg.ChatBanDuration,
		cli.cfg.ChatBanReason,
	)
	defer func() {
		err = errors.Join(err, broker.Close())
	}()

	if len(cli.cfg.IPBlacklists) > 0 {
		log.Println("loading ip blacklists...")
		for _, filePath := range cli.cfg.IPBlacklists {
			err = broker.AddBlacklistCIDRFile(filePath)
			if err != nil {
				return err
			}
		}
	}

	if len(cli.cfg.ChatBlacklists) > 0 {
		log.Println("loading chat blacklists...")
		for _, filePath := range cli.cfg.ChatBlacklists {
			err = broker.AddBlacklistChatFile(filePath)
			if err != nil {
				return err
			}
		}
	}

	log.Println("connecting to econ servers...")
	for idx, addrPort := range cli.cfg.EconServers {
		err = broker.DialTo(
			cli.ctx,
			addrPort,
			cli.cfg.EconPasswords[idx],
		)
		if err != nil {
			return err
		}
	}

	// block until context is done
	log.Println("banserver started successfully")
	<-cli.ctx.Done()
	log.Println("shutting down banserver...")
	return nil
}
