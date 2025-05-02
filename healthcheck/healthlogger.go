package healthcheck

import (
	"encoding/json"
)

// отдельная горутина для выноса логики логирования статусов и информации

type HealthLogStruct struct {
	Upstream string `json:"Upstream"`
	// core.BackendInfo
	Url     string `json:"ProxyURL"`
	Addr    string `json:"Address"`
	IsAlive bool   `json:"IsAlive"`
}

func HealthLog(toLog HealthLogStruct) {
	jsonStr, err := json.Marshal(toLog)
	if err != nil {
		logger.Info("Error occured during logging backend status")
	}
	logger.Info(string(jsonStr))
}

func StartHealthLogger(log <-chan HealthLogStruct) {
	for logstr := range log {
		HealthLog(logstr)
	}
}
