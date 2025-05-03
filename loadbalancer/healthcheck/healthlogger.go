package healthcheck

// отдельная горутина для выноса логики логирования статусов и информации

type HealthLogStruct struct {
	// core.BackendInfo
	Url     string `json:"ProxyURL"`
	Addr    string `json:"Address"`
	IsAlive bool   `json:"IsAlive"`
}

func HealthLog(toLog HealthLogStruct) {
	// jsonStr, err := json.Marshal(toLog)
	// if err != nil {
	// 	logger.Info("Error occured during logging backend status")
	// }
	logger.Info("backend status",
		"Url", toLog.Url,
		"Addr", toLog.Addr,
		"IsAlive", toLog.IsAlive)
}

func StartHealthLogger(log <-chan HealthLogStruct) {
	for logstr := range log {
		HealthLog(logstr)
	}
	logger.Debug("Healthcheck logger closed")
}
