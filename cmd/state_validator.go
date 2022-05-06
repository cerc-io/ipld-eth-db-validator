package cmd

import (
	"context"
	"fmt"

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

	height := viper.GetUint64("validate.block-height")
	if height < 1 {
		logWithCommand.Fatalf("block height cannot be less the 1")
	}

	trail := viper.GetUint64("validate.trail")
	// TODO: add chain config logic here.
	srvc := validator.NewService(cfg.DB, height, trail, nil)

	_, err = srvc.Start(context.Background())
	if err != nil {
		logWithCommand.Fatal(err)
	}

	logWithCommand.Println("state validation complete")
}

func init() {
	rootCmd.AddCommand(stateValidatorCmd)

	stateValidatorCmd.PersistentFlags().String("block-height", "1", "block height to initiate state validation")
	stateValidatorCmd.PersistentFlags().String("trail", "0", "trail of block height to validate")

	_ = viper.BindPFlag("validate.block-height", stateValidatorCmd.PersistentFlags().Lookup("block-height"))
	_ = viper.BindPFlag("validate.trail", stateValidatorCmd.PersistentFlags().Lookup("trail"))
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
