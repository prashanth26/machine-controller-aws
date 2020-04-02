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

// Package driver contains the cloud provider specific implementations to manage machines
package driver

import (
	corev1 "k8s.io/api/core/v1"
)

// Driver is the common interface for creation/deletion of the VMs over different cloud-providers.
type Driver interface {
	Create() (string, string, error)
	Delete(string) error
	GetExisting() (string, error)
	GetVMs(string) (VMs, error)
	GetVolNames([]corev1.PersistentVolumeSpec) ([]string, error)
	GetUserData() string
	SetUserData(string)
}

// VMs maintains a list of VM returned by the provider
// Key refers to the machine-id on the cloud provider
// value refers to the machine-name of the machine object
type VMs map[string]string
