package main

import (
	"github.com/sirupsen/logrus"
	"time"
)

type metrics struct {
	totalRequests      int64
	successfulRequests int64
	failedRequests     int64
	avgResponseTime    time.Duration
	lastReport         time.Time
}

var serverMetrics = &metrics{
	lastReport: time.Now(),
}

func collectMetrics() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		log.WithFields(logrus.Fields{
			"total_requests":       serverMetrics.totalRequests,
			"successful_requests":  serverMetrics.successfulRequests,
			"failed_requests":      serverMetrics.failedRequests,
			"avg_response_time_ms": serverMetrics.avgResponseTime.Milliseconds(),
			"uptime":               time.Since(serverMetrics.lastReport).String(),
		}).Info("Server metrics")
	}
}
