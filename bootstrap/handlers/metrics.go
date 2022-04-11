//
// Copyright (c) 2022 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package handlers

import (
	"context"
	"sync"
	"time"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/metrics"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
)

type RegisterTelemetryFunc func(logger.LoggingClient, *config.TelemetryInfo, interfaces.MetricsManager)

type ServiceMetrics struct {
	serviceName string
}

func NewServiceMetrics(serviceName string) *ServiceMetrics {
	return &ServiceMetrics{
		serviceName: serviceName,
	}
}

// BootstrapHandler fulfills the BootstrapHandler contract and performs initialization of service metrics.
func (s *ServiceMetrics) BootstrapHandler(ctx context.Context, wg *sync.WaitGroup, _ startup.Timer, dic *di.Container) bool {
	lc := container.LoggingClientFrom(dic.Get)
	serviceConfig := container.ConfigurationFrom(dic.Get)

	messageClient := container.MessagingClientFrom(dic.Get)
	telemetryConfig := serviceConfig.GetTelemetryInfo()

	interval, err := time.ParseDuration(telemetryConfig.Interval)
	if err != nil {
		lc.Errorf("Telemetry interval is invalid time duration: %s", err.Error())
		return false
	}

	reporter := metrics.NewMessageBusReporter(lc, s.serviceName, messageClient, telemetryConfig)
	manager := metrics.NewManager(lc, interval, reporter)

	manager.Run(ctx, wg)

	dic.Update(di.ServiceConstructorMap{
		container.MetricsManagerInterfaceName: func(get di.Get) interface{} {
			return manager
		},
	})

	return true
}
