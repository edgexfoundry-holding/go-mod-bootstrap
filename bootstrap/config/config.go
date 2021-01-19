/*******************************************************************************
 * Copyright 2019 Dell Inc.
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

package config

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"
	"sync"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/environment"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/flags"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"

	"github.com/edgexfoundry/go-mod-configuration/v2/configuration"
	"github.com/edgexfoundry/go-mod-configuration/v2/pkg/types"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"

	"github.com/BurntSushi/toml"
)

const writableKey = "/Writable"

// UpdatedStream defines the stream type that is notified by ListenForChanges when a configuration update is received.
type UpdatedStream chan struct{}

type Processor struct {
	Logger          logger.LoggingClient
	flags           flags.Common
	envVars         *environment.Variables
	startupTimer    startup.Timer
	ctx             context.Context
	wg              *sync.WaitGroup
	configUpdated   UpdatedStream
	dic             *di.Container
	overwriteConfig bool
}

// NewProcessor creates a new configuration Processor
func NewProcessor(
	lc logger.LoggingClient,
	flags flags.Common,
	envVars *environment.Variables,
	startupTimer startup.Timer,
	ctx context.Context,
	wg *sync.WaitGroup,
	configUpdated UpdatedStream,
	dic *di.Container,
) *Processor {
	return &Processor{
		Logger:        lc,
		flags:         flags,
		envVars:       envVars,
		startupTimer:  startupTimer,
		ctx:           ctx,
		wg:            wg,
		configUpdated: configUpdated,
		dic:           dic,
	}
}

func (cp *Processor) Process(serviceKey string, configStem string, serviceConfig interfaces.Configuration) error {
	// Create some shorthand for frequently used items
	envVars := cp.envVars
	lc := cp.Logger

	cp.overwriteConfig = cp.flags.OverwriteConfig()

	// Local configuration must be loaded first in case need registry config info and/or
	// need to push it to the Configuration Provider.
	if err := cp.loadFromFile(serviceConfig); err != nil {
		return err
	}

	// Override file-based configuration with envVars variables.
	// Variables variable overrides have precedence over all others,
	// so make sure they are applied before config is used for anything.
	overrideCount, err := envVars.OverrideConfiguration(serviceConfig)
	if err != nil {
		return err
	}

	configProviderUrl := cp.flags.ConfigProviderUrl()

	// Create new ProviderInfo and initialize it from command-line flag or Variables variables
	configProviderInfo, err := NewProviderInfo(cp.envVars, configProviderUrl)
	if err != nil {
		return err
	}

	switch configProviderInfo.UseProvider() {
	case true:
		configClient, err := cp.createProviderClient(serviceKey, configStem, configProviderInfo.ServiceConfig())
		if err != nil {
			return fmt.Errorf("failed to create Configuration Provider client: %s", err.Error())
		}

		for cp.startupTimer.HasNotElapsed() {
			if err := cp.processWithProvider(
				configClient,
				serviceConfig,
				overrideCount,
			); err != nil {
				lc.Error(err.Error())
				select {
				case <-cp.ctx.Done():
					return errors.New("aborted Updating to/from Configuration Provider")
				default:
					cp.startupTimer.SleepForInterval()
					continue
				}
			}

			break
		}

		cp.listenForChanges(serviceConfig, configClient)

	case false:
		cp.logConfigInfo("Using local configuration from file", overrideCount)
	}

	// Now that configuration has been loaded and overrides applied the log level can be set as configured.
	lc.SetLogLevel(serviceConfig.GetLogLevel())

	return err
}

// createProviderClient creates and returns a configuration.Client instance and logs Client connection information
func (cp *Processor) createProviderClient(
	serviceKey string,
	configStem string,
	providerConfig types.ServiceConfig) (configuration.Client, error) {

	providerConfig.BasePath = configStem + serviceKey

	cp.Logger.Info(fmt.Sprintf(
		"Using Configuration provider (%s) from: %s with base path of %s",
		providerConfig.Type,
		providerConfig.GetUrl(),
		providerConfig.BasePath))

	return configuration.NewConfigurationClient(providerConfig)
}

// LoadFromFile attempts to read and unmarshal toml-based configuration into a configuration struct.
func (cp *Processor) loadFromFile(config interfaces.Configuration) error {
	configDir := environment.GetConfDir(cp.Logger, cp.flags.ConfigDirectory())
	profileDir := environment.GetProfileDir(cp.Logger, cp.flags.Profile())
	configFileName := environment.GetConfigFileName(cp.Logger, cp.flags.ConfigFileName())

	filePath := configDir + "/" + profileDir + configFileName

	contents, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("could not load configuration file (%s): %s", filePath, err.Error())
	}
	if err = toml.Unmarshal(contents, config); err != nil {
		return fmt.Errorf("could not load configuration file (%s): %s", filePath, err.Error())
	}

	cp.Logger.Info(fmt.Sprintf("Loaded configuration from %s", filePath))

	return nil
}

// ProcessWithProvider puts configuration if doesnt exist in provider (i.e. self-seed) or
// gets configuration from provider and updates the service's configuration with envVars overrides after receiving
// them from the provider so that envVars override supersede any changes made in the provider.
func (cp *Processor) processWithProvider(
	configClient configuration.Client,
	serviceConfig interfaces.Configuration,
	overrideCount int) error {

	if !configClient.IsAlive() {
		return errors.New("configuration provider is not available")
	}

	hasConfig, err := configClient.HasConfiguration()
	if err != nil {
		return fmt.Errorf("could not determine if Configuration provider has configuration: %s", err.Error())
	}

	if !hasConfig || cp.overwriteConfig {
		// Variables overrides already applied previously so just push to Configuration Provider
		// Note that serviceConfig is a pointer, so we have to use reflection to dereference it.
		err = configClient.PutConfiguration(reflect.ValueOf(serviceConfig).Elem().Interface(), true)
		if err != nil {
			return fmt.Errorf("could not push configuration into Configuration Provider: %s", err.Error())
		}

		cp.logConfigInfo("Configuration has been pushed to into Configuration Provider", overrideCount)
	} else {
		rawConfig, err := configClient.GetConfiguration(serviceConfig)
		if err != nil {
			return fmt.Errorf("could not get configuration from Configuration provider: %s", err.Error())
		}

		if !serviceConfig.UpdateFromRaw(rawConfig) {
			return errors.New("configuration from Configuration provider failed type check")
		}

		overrideCount, err := cp.envVars.OverrideConfiguration(serviceConfig)
		if err != nil {
			return err
		}

		cp.logConfigInfo("Configuration has been pulled from Configuration provider", overrideCount)
	}

	return nil
}

// listenForChanges leverages the Configuration Provider client's WatchForChanges() method to receive changes to and update the
// service's configuration writable sub-struct.  It's assumed the log level is universally part of the
// writable struct and this function explicitly updates the loggingClient's log level when new configuration changes
// are received.
func (cp *Processor) listenForChanges(serviceConfig interfaces.Configuration, configClient configuration.Client) {
	lc := cp.Logger
	isFirstUpdate := true

	cp.wg.Add(1)
	go func() {
		defer cp.wg.Done()

		errorStream := make(chan error)
		defer close(errorStream)

		updateStream := make(chan interface{})
		defer close(updateStream)

		configClient.WatchForChanges(updateStream, errorStream, serviceConfig.EmptyWritablePtr(), writableKey)

		for {
			select {
			case <-cp.ctx.Done():
				return

			case ex := <-errorStream:
				lc.Error(ex.Error())

			case raw, ok := <-updateStream:
				if !ok {
					return
				}

				// Config Provider sends an update as soon as the watcher is connected even though there are not
				// any changes to the configuration. This causes an issue during start-up if there is an
				// envVars override of one of the Writable fields, so we must ignore the first update.
				if isFirstUpdate {
					isFirstUpdate = false
					continue
				}

				previousInsecureSecrets := serviceConfig.GetInsecureSecrets()
				previousLogLevel := serviceConfig.GetLogLevel()

				if !serviceConfig.UpdateWritableFromRaw(raw) {
					lc.Error("ListenForChanges() type check failed")
					return
				}

				currentInsecureSecrets := serviceConfig.GetInsecureSecrets()
				currentLogLevel := serviceConfig.GetLogLevel()

				lc.Info("Writeable configuration has been updated from the Configuration Provider")

				// Note: Updates occur one setting at a time so only have to look for single changes
				switch {
				case currentLogLevel != previousLogLevel:
					_ = lc.SetLogLevel(serviceConfig.GetLogLevel())
					lc.Info(fmt.Sprintf("Logging level changed to %s", currentLogLevel))

				// InsecureSecrets (map) will be nil if not in the original TOML used to seed the Config Provider,
				// so ignore it if this is the case.
				case currentInsecureSecrets != nil &&
					!reflect.DeepEqual(currentInsecureSecrets, previousInsecureSecrets):
					lc.Info("Insecure Secrets have been updated")
					secretProvider := container.SecretProviderFrom(cp.dic.Get)
					if secretProvider != nil {
						secretProvider.SecretsUpdated()
					}

				default:
					// Signal that configuration updates exists that have not already been processed.
					if cp.configUpdated != nil {
						cp.configUpdated <- struct{}{}
					}
				}
			}
		}
	}()
}

// logConfigInfo logs the config info message with number over overrides that occurred.
func (cp *Processor) logConfigInfo(message string, overrideCount int) {
	cp.Logger.Info(fmt.Sprintf("%s (%d envVars overrides applied)", message, overrideCount))
}
