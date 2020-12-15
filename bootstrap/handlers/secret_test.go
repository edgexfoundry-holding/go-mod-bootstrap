/*******************************************************************************
 * Copyright 2020 Intel Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/

package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"sync"
	"testing"

	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/secret"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/startup"
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/config"
	"github.com/edgexfoundry/go-mod-bootstrap/di"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-secrets/pkg/token/authtokenloader/mocks"
	"github.com/edgexfoundry/go-mod-secrets/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	expectedUsername = "admin"
	expectedPassword = "password"
	expectedPath     = "/redisdb"
)

var testTokenResponse = `{"auth":{"accessor":"9OvxnrjgV0JTYMeBreak7YJ9","client_token":"s.oPJ8uuJCkTRb2RDdcNova8wg","entity_id":"","lease_duration":3600,"metadata":{"edgex-service-name":"edgex-core-data"},"orphan":true,"policies":["default","edgex-service-edgex-core-data"],"renewable":true,"token_policies":["default","edgex-service-edgex-core-data"],"token_type":"service"},"data":null,"lease_duration":0,"lease_id":"","renewable":false,"request_id":"ee749ee1-c8bf-6fa9-3ed5-644181fc25b0","warnings":null,"wrap_info":null}`
var expectedSecrets = map[string]string{secret.UsernameKey: expectedUsername, secret.PasswordKey: expectedPassword}

func TestProvider_BootstrapHandler(t *testing.T) {
	tests := []struct {
		Name   string
		Secure string
	}{
		{"Valid Secure", "true"},
		{"Valid Insecure", "false"},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			_ = os.Setenv(secret.EnvSecretStore, tc.Secure)
			timer := startup.NewStartUpTimer("UnitTest")

			dic := di.NewContainer(di.ServiceConstructorMap{
				container.LoggingClientInterfaceName: func(get di.Get) interface{} {
					return logger.MockLogger{}
				},
				container.ConfigurationInterfaceName: func(get di.Get) interface{} {
					return TestConfig{}
				},
			})

			if tc.Secure == "true" {
				testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					switch r.RequestURI {
					case "/v1/auth/token/lookup-self":
						w.WriteHeader(http.StatusOK)
						_, _ = w.Write([]byte(testTokenResponse))
					case "/redisdb":
						w.WriteHeader(http.StatusOK)
						data := make(map[string]interface{})
						data["data"] = expectedSecrets
						response, _ := json.Marshal(data)
						_, _ = w.Write(response)
					default:
						w.WriteHeader(http.StatusNotFound)
					}
				}))
				defer testServer.Close()

				serverUrl, _ := url.Parse(testServer.URL)
				port, _ := strconv.Atoi(serverUrl.Port())
				config := NewTestConfig(port)

				mockTokenLoader := &mocks.AuthTokenLoader{}
				mockTokenLoader.On("Load", "token.json").Return("Test Token", nil)
				dic.Update(di.ServiceConstructorMap{
					container.AuthTokenLoaderInterfaceName: func(get di.Get) interface{} {
						return mockTokenLoader
					},
					container.ConfigurationInterfaceName: func(get di.Get) interface{} {
						return config
					},
				})
			}

			actual := SecureProviderBootstrapHandler(context.Background(), &sync.WaitGroup{}, timer, dic)
			require.True(t, actual)
			actualProvider := container.SecretProviderFrom(dic.Get)
			assert.NotNil(t, actualProvider)

			actualSecrets, err := actualProvider.GetSecrets(expectedPath)
			require.NoError(t, err)
			assert.Equal(t, expectedUsername, actualSecrets[secret.UsernameKey])
			assert.Equal(t, expectedPassword, actualSecrets[secret.PasswordKey])
		})
	}
}

type TestConfig struct {
	InsecureSecrets bootstrapConfig.InsecureSecrets
	SecretStore     bootstrapConfig.SecretStoreInfo
}

func NewTestConfig(port int) TestConfig {
	return TestConfig{
		SecretStore: bootstrapConfig.SecretStoreInfo{
			Host:       "localhost",
			Port:       port,
			Protocol:   "http",
			ServerName: "localhost",
			TokenFile:  "token.json",
			Authentication: types.AuthenticationInfo{
				AuthType:  "Dummy-Token",
				AuthToken: "myToken",
			},
		},
	}
}

func (t TestConfig) UpdateFromRaw(_ interface{}) bool {
	panic("implement me")
}

func (t TestConfig) EmptyWritablePtr() interface{} {
	panic("implement me")
}

func (t TestConfig) UpdateWritableFromRaw(_ interface{}) bool {
	panic("implement me")
}

func (t TestConfig) GetBootstrap() bootstrapConfig.BootstrapConfiguration {
	return bootstrapConfig.BootstrapConfiguration{
		SecretStore: t.SecretStore,
	}
}

func (t TestConfig) GetLogLevel() string {
	panic("implement me")
}

func (t TestConfig) GetRegistryInfo() bootstrapConfig.RegistryInfo {
	panic("implement me")
}

func (t TestConfig) GetInsecureSecrets() bootstrapConfig.InsecureSecrets {
	return map[string]bootstrapConfig.InsecureSecretsInfo{
		"DB": {
			Path:    expectedPath,
			Secrets: expectedSecrets,
		},
	}
}
