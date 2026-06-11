package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
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

	cmd := &cobra.Command{
		Use:   constants.CommandRun,
		Short: "Run TraceDeck agent",
		RunE: func(cmd *cobra.Command, _ []string) error {
			policy, err := config.LoadFile(configPath)
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "TraceDeck agent bootstrap ok: tenant=%s device=%s profile=%s\n", policy.TenantID, policy.DeviceID, policy.Profile)
			return err
		},
	}
	cmd.Flags().StringVar(&configPath, "config", constants.DefaultConfig, "policy YAML path")
	return cmd
}
