//
// Copyright (c) 2020 Intel Corporation
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
//

package registration

import (
	"testing"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/config"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateRegistryClient(t *testing.T) {
	lc := logger.NewMockClient()

	serviceConfig := unitTestConfiguration{
		Service: config.ServiceInfo{
			Host:     "localhost",
			Port:     8080,
			Protocol: "http",
		},
		Registry: config.RegistryInfo{
			Host: "localhost",
			Port: 8500,
			Type: "consul",
		},
	}

	actual, err := createRegistryClient("unit-test", serviceConfig, lc)
	require.NoError(t, err)
	assert.NotNil(t, actual)
}

type unitTestConfiguration struct {
	Service  config.ServiceInfo
	Registry config.RegistryInfo
}

func (ut unitTestConfiguration) GetInsecureSecrets() config.InsecureSecrets {
	return nil
}

func (ut unitTestConfiguration) UpdateFromRaw(_ interface{}) bool {
	panic("should not be called")
}

func (ut unitTestConfiguration) EmptyWritablePtr() interface{} {
	panic("should not be called")
}

func (ut unitTestConfiguration) UpdateWritableFromRaw(_ interface{}) bool {
	panic("should not be called")
}

func (ut unitTestConfiguration) GetBootstrap() config.BootstrapConfiguration {
	return config.BootstrapConfiguration{
		Service:  ut.Service,
		Registry: ut.Registry,
	}
}

func (ut unitTestConfiguration) GetLogLevel() string {
	return "TRACE"
}

func (ut unitTestConfiguration) GetRegistryInfo() config.RegistryInfo {
	return ut.Registry
}
