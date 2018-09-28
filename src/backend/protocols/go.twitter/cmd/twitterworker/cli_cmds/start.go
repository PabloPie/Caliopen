/*
 * // Copyleft (ɔ) 2018 The Caliopen contributors.
 * // Use of this source code is governed by a GNU AFFERO GENERAL PUBLIC
 * // license (AGPL) that can be found in the LICENSE file.
 */

/*
 * // Copyleft (ɔ) 2018 The Caliopen contributors.
 * // Use of this source code is governed by a GNU AFFERO GENERAL PUBLIC
 * // license (AGPL) that can be found in the LICENSE file.
 */

package cmd

import (
	"fmt"
	twd "github.com/CaliOpen/Caliopen/src/backend/protocols/go.twitter"
	log "github.com/Sirupsen/logrus"
	"github.com/nats-io/go-nats"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	configPath     string
	configFile     string
	pidFile        string
	signalChannel  chan os.Signal


	startCmd = &cobra.Command{
		Use:   "start",
		Short: "Starts a pool of twitter API worker(s)",
		Run:   start,
	}
)

func init() {
	startCmd.PersistentFlags().StringVarP(&configFile, "config", "c",
		"twitterworker", "Name of the configuration file, without extension. (YAML, TOML, JSON… allowed)")
	startCmd.PersistentFlags().StringVarP(&configPath, "configpath", "",
		"../../../../configs/", "Main config file path.")
	startCmd.PersistentFlags().StringVarP(&pidFile, "pid-file", "p",
		"/var/run/caliopen_twitterd.pid", "Path to the pid file")

	RootCmd.AddCommand(startCmd)
	signalChannel = make(chan os.Signal, 1)
}

func start(cmd *cobra.Command, args []string) {

	err := readConfig(twd.AppConfig)
	if err != nil {
		log.WithError(err).Fatal("Error while reading config")
	}
	// Write out our PID
	if len(pidFile) > 0 {
		if f, err := os.Create(pidFile); err == nil {
			defer f.Close()
			if _, err := f.WriteString(fmt.Sprintf("%d", os.Getpid())); err == nil {
				f.Sync()
			} else {
				log.WithError(err).Warnf("Error while writing pidFile (%s)", pidFile)
			}
		} else {
			log.WithError(err).Warnf("Error while creating pidFile (%s)", pidFile)
		}
	}

	twd.WorkersGuard = new(sync.RWMutex)
	twd.TwitterWorkers = []*twd.Worker{}
	//TODO: listen to NATS messages
	twd.NatsConn, err = nats.Connect(twd.AppConfig.BrokerConfig.NatsURL)
	if err != nil {
		log.WithError(err).Fatal("[TwitterWorker] initialization of NATS connexion failed")
	}
	_, err = twd.NatsConn.QueueSubscribe(twd.AppConfig.BrokerConfig.NatsTopicFetcher, twd.AppConfig.BrokerConfig.NatsQueue, twd.NatsMsgHandler)
	if err != nil {
		log.WithError(err).Fatal("[TwitterWorker] initialization of NATS subscription failed")
	}

	/*
	// retrieve twitter remote identities from db
	accounts, err := twd.RetrieveTwitterAccounts()
	if err != nil {
		log.WithError(err).Fatal("failed to retrieve twitter credentials from db")
	}

	// init and start worker(s) for each identity
	twitterWorkers = []*twd.Worker{}
	var i int
	for _, account := range accounts {
		log.Infof("initializing Twitter worker %d", i)
		worker, err := twd.NewWorker(i, twd.AppConfig, account, twd.DefaultPollInterval, "")
		if err != nil {
			log.WithError(err).Warnf("Failed to init twitter Worker %d", i)
		} else {
			twitterWorkers = append(twitterWorkers, worker)
			go twitterWorkers[i].Start()
			i++
		}
	}
	workersGuard = new(sync.RWMutex)
	go sigHandler(twitterWorkers)



	/*
	// start polling
	ticker := time.NewTicker(twd.DefaultPollInterval * time.Second)
	for range ticker.C {
		workersGuard.RLock()
		for _, worker := range twitterWorkers {
			if worker.IsDue() {
				worker.PollDesk <- twd.PollDM
			}
		}
		workersGuard.RUnlock()
	}
*/
}

// ReadConfig which should be called at startup, or when a SIG_HUP is caught
func readConfig(config *twd.WorkerConfig) error {
	// load in the main config. Reading from YAML, TOML, JSON, HCL and Java properties config files
	v := viper.New()
	v.SetConfigName(configFile)                           // name of config file (without extension)
	v.AddConfigPath(configPath)                           // path to look for the config file in
	v.AddConfigPath("$CALIOPENROOT/src/backend/configs/") // call multiple times to add many search paths
	v.AddConfigPath(".")                                  // optionally look for config in the working directory

	err := v.ReadInConfig() // Find and read the config file*/
	if err != nil {
		log.WithError(err).Infof("Could not read main config file <%s>.", configFile)
		return err
	}
	err = v.Unmarshal(config)
	if err != nil {
		log.WithError(err).Infof("Could not parse config file: <%s>", configFile)
		return err
	}

	return nil
}
func sigHandler(workers []*twd.Worker) {
	signal.Notify(signalChannel, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGKILL)

	for sig := range signalChannel {

		if sig == syscall.SIGHUP {
			err := readConfig(twd.AppConfig)
			if err != nil {
				log.WithError(err).Error("Error while ReadConfig (reload)")
			} else {
				log.Info("Configuration is reloaded")
			}
			// TODO: handle SIGHUP
		} else if sig == syscall.SIGTERM || sig == syscall.SIGQUIT || sig == syscall.SIGINT {
			log.Info("Shutdown signal caught")
			for _, w := range workers {
				w.PollDesk <- twd.Stop
			}
			log.Info("Shutdown completed, exiting")
			os.Exit(0)
		} else {
			os.Exit(0)
		}
	}
}
