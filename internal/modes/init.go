package modes

import (
	"fmt"
	"os"

	"ssh_connect/internal/cli"
	"ssh_connect/internal/config"
)

func Init(opts cli.Options) error {
	if _, err := os.Stat(opts.ConfigPath); err == nil {
		return fmt.Errorf("config file already exists: %s", opts.ConfigPath)
	}

	cfg := config.ExampleConfig()
	if err := config.Save(opts.ConfigPath, cfg); err != nil {
		return err
	}

	fmt.Printf("Example config created: %s\n", opts.ConfigPath)
	return nil
}
