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
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cerc-io/ipld-eth-db-validator/v5/pkg/prom"
)

var (
	cfgFile        string
	subCommand     string
	logWithCommand log.Entry
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:              "ipld-eth-db-validator",
	Short:            "Validates each block state stored for state-diff service.",
	Long:             `Validates each block state stored for state-diff service.`,
	PersistentPreRun: initFunc,
}

func Execute() {
	log.Info("----- Starting state validator -----")
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func initFunc(cmd *cobra.Command, args []string) {
	logfile := viper.GetString("log.file")
	if logfile != "" {
		file, err := os.OpenFile(logfile,
			os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			log.Infof("Directing output to %s", logfile)
			log.SetOutput(file)
		} else {
			log.SetOutput(os.Stdout)
			log.Info("Failed to log to file, using default stdout")
		}
	} else {
		log.SetOutput(os.Stdout)
	}

	lvl, err := log.ParseLevel(viper.GetString("log.level"))
	if err != nil {
		log.Fatal("Could not parse log level: ", err)
	}
	log.SetLevel(lvl)
	if lvl > log.InfoLevel {
		log.SetReportCaller(true)
	}
	log.Info("Log level set to ", lvl)

	if viper.GetBool("prom.metrics") {
		log.Info("initializing prometheus metrics")
		prom.Init()
	}

	if viper.GetBool("prom.http") {
		addr := fmt.Sprintf(
			"%s:%s",
			viper.GetString("prom.httpAddr"),
			viper.GetString("prom.httpPort"),
		)
		log.Info("starting prometheus server")
		prom.Serve(addr)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file location")
	rootCmd.PersistentFlags().String("database-name", "cerc_public", "database name")
	rootCmd.PersistentFlags().Int("database-port", 5432, "database port")
	rootCmd.PersistentFlags().String("database-hostname", "localhost", "database hostname")
	rootCmd.PersistentFlags().String("database-user", "", "database user")
	rootCmd.PersistentFlags().String("database-password", "", "database password")
	rootCmd.PersistentFlags().String("log-file", "", "file path for logging")
	rootCmd.PersistentFlags().String("log-level", log.InfoLevel.String(), "Log level (trace, debug, info, warn, error, fatal, panic")

	rootCmd.PersistentFlags().Bool("prom-metrics", false, "enable prometheus metrics")
	rootCmd.PersistentFlags().Bool("prom-http", false, "enable prometheus http service")
	rootCmd.PersistentFlags().String("prom-httpAddr", "127.0.0.1", "prometheus http host")
	rootCmd.PersistentFlags().String("prom-httpPort", "9001", "prometheus http port")
	rootCmd.PersistentFlags().Bool("prom-dbStats", false, "enables prometheus db stats")

	_ = viper.BindPFlag("database.name", rootCmd.PersistentFlags().Lookup("database-name"))
	_ = viper.BindPFlag("database.port", rootCmd.PersistentFlags().Lookup("database-port"))
	_ = viper.BindPFlag("database.hostname", rootCmd.PersistentFlags().Lookup("database-hostname"))
	_ = viper.BindPFlag("database.user", rootCmd.PersistentFlags().Lookup("database-user"))
	_ = viper.BindPFlag("database.password", rootCmd.PersistentFlags().Lookup("database-password"))
	_ = viper.BindPFlag("log.file", rootCmd.PersistentFlags().Lookup("log-file"))
	_ = viper.BindPFlag("log.level", rootCmd.PersistentFlags().Lookup("log-level"))

	_ = viper.BindPFlag("prom.metrics", rootCmd.PersistentFlags().Lookup("prom-metrics"))
	_ = viper.BindPFlag("prom.http", rootCmd.PersistentFlags().Lookup("prom-http"))
	_ = viper.BindPFlag("prom.httpAddr", rootCmd.PersistentFlags().Lookup("prom-httpAddr"))
	_ = viper.BindPFlag("prom.httpPort", rootCmd.PersistentFlags().Lookup("prom-httpPort"))
	_ = viper.BindPFlag("prom.dbStats", rootCmd.PersistentFlags().Lookup("prom-dbStats"))
}
