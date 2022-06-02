package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/vulcanize/ipld-eth-db-validator/pkg/validator"
)

// stateValidatorCmd represents the stateValidator command
var stateValidatorCmd = &cobra.Command{
	Use:   "stateValidator",
	Short: "Validate ethereum state",
	Long:  `Usage ./ipld-eth-db-validator stateValidator --config={path to toml config file}`,

	Run: func(cmd *cobra.Command, args []string) {
		subCommand = cmd.CalledAs()
		logWithCommand = *log.WithField("SubCommand", subCommand)
		stateValidator()
	},
}

func stateValidator() {
	cfg, err := validator.NewConfig()
	if err != nil {
		logWithCommand.Fatal(err)
	}

	service := validator.NewService(cfg, nil)

	wg := new(sync.WaitGroup)
	wg.Add(1)
	go service.Start(context.Background(), wg)

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt)
	<-shutdown
	service.Stop()
	wg.Wait()
}

func init() {
	rootCmd.AddCommand(stateValidatorCmd)

	stateValidatorCmd.PersistentFlags().String("block-height", "1", "block height to initiate state validation")
	stateValidatorCmd.PersistentFlags().String("trail", "16", "trail of block height to validate")
	stateValidatorCmd.PersistentFlags().String("sleep-interval", "10", "sleep interval in seconds after validator has caught up to (head-trail) height")

	stateValidatorCmd.PersistentFlags().String("chain-config", "", "path to chain config")

	_ = viper.BindPFlag("validate.block-height", stateValidatorCmd.PersistentFlags().Lookup("block-height"))
	_ = viper.BindPFlag("validate.trail", stateValidatorCmd.PersistentFlags().Lookup("trail"))
	_ = viper.BindPFlag("validate.sleepInterval", stateValidatorCmd.PersistentFlags().Lookup("sleep-interval"))

	_ = viper.BindPFlag("ethereum.chainConfig", stateValidatorCmd.PersistentFlags().Lookup("chain-config"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
		if err := viper.ReadInConfig(); err == nil {
			log.Printf("Using config file: %s", viper.ConfigFileUsed())
		} else {
			log.Fatal(fmt.Sprintf("Couldn't read config file: %s", err.Error()))
		}
	} else {
		log.Warn("No config file passed with --config flag")
	}
}
