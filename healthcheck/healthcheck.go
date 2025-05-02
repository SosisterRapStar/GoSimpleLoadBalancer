package healthcheck

import (
	"context"

	"github.com/SosisterRapStar/GoSimpleLoadBalancer/core"
)

// Описание интерфейса чекера и обычный используемый чекер

type HealthChecker interface {
	StartCheck(ctx context.Context, updates <-chan *core.HealthStatus)
}
