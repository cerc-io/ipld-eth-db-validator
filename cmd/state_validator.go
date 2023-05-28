// VulcanizeDB
// Copyright © 2022 Vulcanize

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.

// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

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

	"github.com/cerc-io/ipld-eth-db-validator/v5/pkg/validator"
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

	service, err := validator.NewService(cfg, nil)
	if err != nil {
		logWithCommand.Fatal(err)
	}

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

	stateValidatorCmd.PersistentFlags().String("from-block", "1", "block height to initiate state validation")
	stateValidatorCmd.PersistentFlags().String("trail", "16", "trail of block height to validate")
	stateValidatorCmd.PersistentFlags().String("retry-interval", "10s", "retry interval in seconds after validator has caught up to (head-trail) height")
	stateValidatorCmd.PersistentFlags().Bool("statediff-missing-block", false, "whether to perform a statediffing call on a missing block")
	stateValidatorCmd.PersistentFlags().String("statediff-timeout", "240s", "statediffing call timeout period (in sec)")

	stateValidatorCmd.PersistentFlags().String("eth-chain-config", "", "path to json chain config")
	stateValidatorCmd.PersistentFlags().String("eth-chain-id", "1", "eth chain id")
	stateValidatorCmd.PersistentFlags().String("eth-http-path", "", "http url for a statediffing node")

	_ = viper.BindPFlag("validate.fromBlock", stateValidatorCmd.PersistentFlags().Lookup("from-block"))
	_ = viper.BindPFlag("validate.trail", stateValidatorCmd.PersistentFlags().Lookup("trail"))
	_ = viper.BindPFlag("validate.retryInterval", stateValidatorCmd.PersistentFlags().Lookup("retry-interval"))
	_ = viper.BindPFlag("validate.stateDiffMissingBlock", stateValidatorCmd.PersistentFlags().Lookup("statediff-missing-block"))
	_ = viper.BindPFlag("validate.stateDiffTimeout", stateValidatorCmd.PersistentFlags().Lookup("statediff-timeout"))

	_ = viper.BindPFlag("ethereum.chainConfig", stateValidatorCmd.PersistentFlags().Lookup("eth-chain-config"))
	_ = viper.BindPFlag("ethereum.chainID", stateValidatorCmd.PersistentFlags().Lookup("eth-chain-id"))
	_ = viper.BindPFlag("ethereum.httpPath", stateValidatorCmd.PersistentFlags().Lookup("eth-http-path"))
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
