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
	var version string

	cmd := &cobra.Command{
		Use:   constants.CommandSchema,
		Short: "Generate TraceDeck policy JSON Schema",
		RunE: func(cmd *cobra.Command, _ []string) error {
			policyVersion, err := schema.ParsePolicyVersion(version)
			if err != nil {
				return err
			}
			data, err := schema.GeneratePolicy(policyVersion)
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
	cmd.Flags().StringVar(&version, "version", string(schema.LatestPolicyVersion()), "policy schema version")
	return cmd
}

func newRunCommand() *cobra.Command {
	var configPath string
	var dataDir string
	var logDir string
	var logLevel string
	var outboxDir string
	var once bool
	var processLimit int
	var archiveOnce bool
	var archiveDryRun bool
	var alertOnce bool
	var alertDryRun bool
	var collectionInterval string
	var maxCycles int
	var browserHistoryPath []string
	var browserHistoryLimit int
	var browserCacheDir string
	var softwareCacheDir string
	var disableBrowserHistory bool

	cmd := &cobra.Command{
		Use:   constants.CommandRun,
		Short: "Run TraceDeck agent",
		RunE: func(cmd *cobra.Command, _ []string) error {
			result, err := app.Run(context.Background(), app.RunOptions{
				ConfigPath:            configPath,
				DataDir:               dataDir,
				LogDir:                logDir,
				LogLevel:              logLevel,
				OutboxDir:             outboxDir,
				Once:                  once,
				ProcessLimit:          processLimit,
				ArchiveOnce:           archiveOnce,
				ArchiveDryRun:         archiveDryRun,
				AlertOnce:             alertOnce,
				AlertDryRun:           alertDryRun,
				CollectionInterval:    collectionInterval,
				MaxCycles:             maxCycles,
				BrowserHistoryPath:    browserHistoryPath,
				BrowserHistoryLimit:   browserHistoryLimit,
				BrowserCacheDir:       browserCacheDir,
				SoftwareCacheDir:      softwareCacheDir,
				DisableBrowserHistory: disableBrowserHistory,
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
	cmd.Flags().StringVar(&outboxDir, "outbox-dir", constants.DefaultOutboxDir, "local archive and alert outbox directory")
	cmd.Flags().BoolVar(&once, "once", false, "collect one local snapshot and exit")
	cmd.Flags().IntVar(&processLimit, "process-limit", constants.DefaultProcessLimit, "maximum processes to persist in one snapshot")
	cmd.Flags().BoolVar(&archiveOnce, "archive-once", false, "stage one archive batch for the collected snapshot")
	cmd.Flags().BoolVar(&archiveDryRun, "archive-dry-run", true, "skip S3 upload after staging the archive batch")
	cmd.Flags().BoolVar(&alertOnce, "alert-once", false, "evaluate alerts for the collected snapshot")
	cmd.Flags().BoolVar(&alertDryRun, "alert-dry-run", true, "write alert notifications to local outbox instead of sending email")
	cmd.Flags().StringVar(&collectionInterval, "collection-interval", constants.DefaultCollectionInterval, "continuous mode collection interval")
	cmd.Flags().IntVar(&maxCycles, "max-cycles", constants.DefaultMaxCycles, "maximum continuous cycles before exit; 0 runs until interrupted")
	cmd.Flags().StringArrayVar(&browserHistoryPath, "browser-history-path", nil, "browser history SQLite path; repeat for multiple paths")
	cmd.Flags().IntVar(&browserHistoryLimit, "browser-history-limit", constants.DefaultBrowserLimit, "maximum browser history rows to inspect per history file")
	cmd.Flags().StringVar(&browserCacheDir, "browser-cache-dir", "", "local browser history cache directory")
	cmd.Flags().StringVar(&softwareCacheDir, "software-cache-dir", "", "local software inventory snapshot cache directory")
	cmd.Flags().BoolVar(&disableBrowserHistory, "disable-browser-history", false, "disable browser history collection for controlled smokes")
	return cmd
}
