package cmd

import (
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	logfile := viper.GetString("logfile")
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

	if err := logLevel(); err != nil {
		log.Fatal("Could not set log level: ", err)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file location")
	rootCmd.PersistentFlags().String("logfile", "", "file path for logging")
	rootCmd.PersistentFlags().String("database-name", "vulcanize_public", "database name")
	rootCmd.PersistentFlags().Int("database-port", 5432, "database port")
	rootCmd.PersistentFlags().String("database-hostname", "localhost", "database hostname")
	rootCmd.PersistentFlags().String("database-user", "", "database user")
	rootCmd.PersistentFlags().String("database-password", "", "database password")
	rootCmd.PersistentFlags().String("log-level", log.InfoLevel.String(), "Log level (trace, debug, info, warn, error, fatal, panic")

	_ = viper.BindPFlag("logfile", rootCmd.PersistentFlags().Lookup("logfile"))
	_ = viper.BindPFlag("database.name", rootCmd.PersistentFlags().Lookup("database-name"))
	_ = viper.BindPFlag("database.port", rootCmd.PersistentFlags().Lookup("database-port"))
	_ = viper.BindPFlag("database.hostname", rootCmd.PersistentFlags().Lookup("database-hostname"))
	_ = viper.BindPFlag("database.user", rootCmd.PersistentFlags().Lookup("database-user"))
	_ = viper.BindPFlag("database.password", rootCmd.PersistentFlags().Lookup("database-password"))
	_ = viper.BindPFlag("log.level", rootCmd.PersistentFlags().Lookup("log-level"))
}

func logLevel() error {
	lvl, err := log.ParseLevel(viper.GetString("log.level"))
	if err != nil {
		return err
	}
	log.SetLevel(lvl)
	if lvl > log.InfoLevel {
		log.SetReportCaller(true)
	}
	log.Info("Log level set to ", lvl.String())
	return nil
}
