/**
 * File              : prometheus.go
 * Author            : Alexandre Saison <alexandre.saison@inarix.com>
 * Date              : 10.02.2021
 * Last Modified Date: 11.02.2021
 * Last Modified By  : Alexandre Saison <alexandre.saison@inarix.com>
 */
package server

import (
	"log"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	migrationLaunched = promauto.NewCounter(prometheus.CounterOpts{
		Name: "goapp_migration_launched",
		Help: "Number of launched migrations",
	})

	seedLaunched = promauto.NewCounter(prometheus.CounterOpts{
		Name: "goapp_seed_launched",
		Help: "Number of launched seeds",
	})

	averageJobTime = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "goapp_avg_job_time",
		Help: "Average go-feather-slack-app launched jobs times",
	})

	metricsChannel = make(chan prometheus.Metric)
)

func (self *Server) recordMetrics() {
	go func() {
		log.Print("Enter goroutine")
		for {
			payload := <-metricsChannel
			value := payload.Desc().String()
			log.Printf("value=%s", value)
			time.Sleep(2 * time.Second)
		}
	}()
}

func (self *Server) increaseMigrationLaunched() {
	migrationLaunched.Inc()
}

func (self *Server) increaseSeedLaunched() {
	seedLaunched.Inc()
}

func (self *Server) updateAvgJobTime() {
	averageJobTime.Collect(metricsChannel)
}
