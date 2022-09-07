// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import (
	time "time"

	mock "github.com/stretchr/testify/mock"
)

// SecretProvider is an autogenerated mock type for the SecretProvider type
type SecretProvider struct {
	mock.Mock
}

// GetAccessToken provides a mock function with given fields: tokenType, serviceKey
func (_m *SecretProvider) GetAccessToken(tokenType string, serviceKey string) (string, error) {
	ret := _m.Called(tokenType, serviceKey)

	var r0 string
	if rf, ok := ret.Get(0).(func(string, string) string); ok {
		r0 = rf(tokenType, serviceKey)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(tokenType, serviceKey)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetSecret provides a mock function with given fields: path, keys
func (_m *SecretProvider) GetSecret(path string, keys ...string) (map[string]string, error) {
	_va := make([]interface{}, len(keys))
	for _i := range keys {
		_va[_i] = keys[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, path)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 map[string]string
	if rf, ok := ret.Get(0).(func(string, ...string) map[string]string); ok {
		r0 = rf(path, keys...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]string)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, ...string) error); ok {
		r1 = rf(path, keys...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListSecretsAtPath provides a mock function with given fields: path
func (_m *SecretProvider) ListSecretsAtPath(path string) ([]string, error) {
	ret := _m.Called(path)

	var r0 []string
	if rf, ok := ret.Get(0).(func(string) []string); ok {
		r0 = rf(path)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(path)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// HasSecret provides a mock function with given fields: path
func (_m *SecretProvider) HasSecret(path string) (bool, error) {
	ret := _m.Called(path)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string) bool); ok {
		r0 = rf(path)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(path)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SecretsLastUpdated provides a mock function with given fields:
func (_m *SecretProvider) SecretsLastUpdated() time.Time {
	ret := _m.Called()

	var r0 time.Time
	if rf, ok := ret.Get(0).(func() time.Time); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(time.Time)
	}

	return r0
}

// SecretsUpdated provides a mock function with given fields:
func (_m *SecretProvider) SecretsUpdated() {
	_m.Called()
}

// StoreSecret provides a mock function with given fields: path, secrets
func (_m *SecretProvider) StoreSecret(path string, secrets map[string]string) error {
	ret := _m.Called(path, secrets)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, map[string]string) error); ok {
		r0 = rf(path, secrets)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewSecretProvider interface {
	mock.TestingT
	Cleanup(func())
}

// NewSecretProvider creates a new instance of SecretProvider. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewSecretProvider(t mockConstructorTestingTNewSecretProvider) *SecretProvider {
	mock := &SecretProvider{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
