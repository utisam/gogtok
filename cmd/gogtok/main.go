package main

import (
	"github.com/sirupsen/logrus"
	"github.com/utisam/gogtok/command"
)

func main() {
	formatter := &logrus.TextFormatter{}
	formatter.DisableTimestamp = true
	logrus.SetFormatter(formatter)

	cmd := command.New()
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	if err := cmd.Execute(); err != nil {
		logrus.WithError(err).Fatal()
	}
}
