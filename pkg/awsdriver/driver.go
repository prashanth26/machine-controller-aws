/*
Copyright (c) 2017 SAP SE or an SAP affiliate company. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package awsdriver contains the cloud provider specific implementations to manage machines
package awsdriver

import (
	"github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha2"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/driver"
	corev1 "k8s.io/api/core/v1"
)

// NewDriver creates a new driver object based on the classKind
func NewDriver(machineID string, secretRef *corev1.Secret, classKind string, machineClass interface{}, machineName string) driver.Driver {

	switch classKind {

	case "AWSMachineClass":
		return &AWSDriver{
			AWSMachineClass: machineClass.(*v1alpha2.AWSMachineClass),
			CloudConfig:     secretRef,
			UserData:        string(secretRef.Data["userData"]),
			MachineID:       machineID,
			MachineName:     machineName,
		}
	}

	return driver.NewFakeDriver(
		func() (string, string, error) {
			return "fake", "fake_ip", nil
		},
		func(string) error {
			return nil
		},
		func() (string, error) {
			return "fake", nil
		},
		func() (driver.VMs, error) {
			return nil, nil
		},
	)
}
