package cmd

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func ParseLogFlags() {
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
}
