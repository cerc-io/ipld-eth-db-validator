package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/params"
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

	chainConfigPath := viper.GetString("ethereum.chainConfig")
	chainCfg, err := LoadConfig(chainConfigPath)
	fmt.Println("chainCfg", chainCfg)
	if err != nil {
		logWithCommand.Fatal(err)
	}

	srvc := validator.NewService(cfg.DB, height, trail, chainCfg)

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

	stateValidatorCmd.PersistentFlags().String("chainConfig", "", "path to chain config")

	_ = viper.BindPFlag("validate.block-height", stateValidatorCmd.PersistentFlags().Lookup("block-height"))
	_ = viper.BindPFlag("validate.trail", stateValidatorCmd.PersistentFlags().Lookup("trail"))

	_ = viper.BindPFlag("ethereum.chainConfig", stateValidatorCmd.PersistentFlags().Lookup("chainConfig"))
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

// LoadConfig loads chain config from json file
func LoadConfig(chainConfigPath string) (*params.ChainConfig, error) {
	file, err := os.Open(chainConfigPath)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to read chain config file: %v", err))

		return nil, err
	}
	defer file.Close()

	chainConfig := new(params.ChainConfig)
	if err := json.NewDecoder(file).Decode(chainConfig); err != nil {
		log.Error(fmt.Sprintf("invalid chain config file: %v", err))

		return nil, err
	}

	log.Info(fmt.Sprintf("Using chain config from %s file. Content %+v", chainConfigPath, chainConfig))

	return chainConfig, nil
}
