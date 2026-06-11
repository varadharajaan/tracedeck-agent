package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/app"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/config"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/constants"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/schema"
)

func NewRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:           constants.AppName,
		Short:         "TraceDeck endpoint observability agent",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.AddCommand(newValidateConfigCommand())
	root.AddCommand(newSchemaCommand())
	root.AddCommand(newRunCommand())
	return root
}

func newValidateConfigCommand() *cobra.Command {
	var configPath string

	cmd := &cobra.Command{
		Use:   constants.CommandValidateConfig,
		Short: "Validate a TraceDeck policy file",
		RunE: func(cmd *cobra.Command, _ []string) error {
			policy, err := config.LoadFile(configPath)
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "valid policy: tenant=%s device=%s profile=%s\n", policy.TenantID, policy.DeviceID, policy.Profile)
			return err
		},
	}
	cmd.Flags().StringVar(&configPath, "config", constants.DefaultConfig, "policy YAML path")
	return cmd
}

func newSchemaCommand() *cobra.Command {
	var outputPath string

	cmd := &cobra.Command{
		Use:   constants.CommandSchema,
		Short: "Generate TraceDeck policy JSON Schema",
		RunE: func(cmd *cobra.Command, _ []string) error {
			data, err := schema.GeneratePolicy(schema.PolicySchemaV1Alpha1)
			if err != nil {
				return err
			}
			if outputPath == "" {
				_, err = cmd.OutOrStdout().Write(data)
				return err
			}
			if err := os.MkdirAll(filepath.Dir(outputPath), 0o750); err != nil {
				return err
			}
			return os.WriteFile(outputPath, data, 0o600)
		},
	}
	cmd.Flags().StringVar(&outputPath, "out", "", "schema output path")
	return cmd
}

func newRunCommand() *cobra.Command {
	var configPath string
	var dataDir string
	var logDir string
	var logLevel string
	var once bool
	var processLimit int

	cmd := &cobra.Command{
		Use:   constants.CommandRun,
		Short: "Run TraceDeck agent",
		RunE: func(cmd *cobra.Command, _ []string) error {
			result, err := app.Run(context.Background(), app.RunOptions{
				ConfigPath:   configPath,
				DataDir:      dataDir,
				LogDir:       logDir,
				LogLevel:     logLevel,
				Once:         once,
				ProcessLimit: processLimit,
			})
			if err != nil {
				return err
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), app.FormatRunResult(result))
			return err
		},
	}
	cmd.Flags().StringVar(&configPath, "config", constants.DefaultConfig, "policy YAML path")
	cmd.Flags().StringVar(&dataDir, "data-dir", constants.DefaultDataDir, "local data directory")
	cmd.Flags().StringVar(&logDir, "log-dir", constants.DefaultLogDir, "local log directory")
	cmd.Flags().StringVar(&logLevel, "log-level", constants.DefaultLogLevel, "log level: trace, debug, info, warn, error")
	cmd.Flags().BoolVar(&once, "once", false, "collect one local snapshot and exit")
	cmd.Flags().IntVar(&processLimit, "process-limit", constants.DefaultProcessLimit, "maximum processes to persist in one snapshot")
	return cmd
}
