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

// FakeDriver is a fake driver returned when none of the actual drivers match
type FakeDriver struct {
	createMachine    func(*CreateMachineRequest) (*CreateMachineResponse, error)
	deleteMachine    func(*DeleteMachineRequest) (*DeleteMachineResponse, error)
	getMachineStatus func(*GetMachineStatusRequest) (*GetMachineStatusResponse, error)
	listMachines     func(*ListMachinesRequest) (*ListMachinesResponse, error)
	getVolumeIDs     func(*GetVolumeIDsRequest) (*GetVolumeIDsResponse, error)
}

// NewFakeDriver returns a new fakedriver object
func NewFakeDriver(
	createMachine func(*CreateMachineRequest) (*CreateMachineResponse, error),
	deleteMachine func(*DeleteMachineRequest) (*DeleteMachineResponse, error),
	getMachineStatus func(*GetMachineStatusRequest) (*GetMachineStatusResponse, error),
	listMachines func(*ListMachinesRequest) (*ListMachinesResponse, error),
	getVolumeIDs func(*GetVolumeIDsRequest) (*GetVolumeIDsResponse, error),
) Driver {
	return &FakeDriver{
		createMachine:    createMachine,
		deleteMachine:    deleteMachine,
		getMachineStatus: getMachineStatus,
		listMachines:     listMachines,
		getVolumeIDs:     getVolumeIDs,
	}
}

// CreateMachine returns a newly created fake driver
func (d *FakeDriver) CreateMachine(createMachineRequest *CreateMachineRequest) (*CreateMachineResponse, error) {
	return d.createMachine(createMachineRequest)
}

// DeleteMachine deletes the fake machine
func (d *FakeDriver) DeleteMachine(deleteMachineRequest *DeleteMachineRequest) (*DeleteMachineResponse, error) {
	return d.deleteMachine(deleteMachineRequest)
}

// GetMachineStatus returns the machine status
func (d *FakeDriver) GetMachineStatus(getMachineStatusRequest *GetMachineStatusRequest) (*GetMachineStatusResponse, error) {
	return d.getMachineStatus(getMachineStatusRequest)
}

// ListMachines returns the list of VMs for a given machineClass
func (d *FakeDriver) ListMachines(listMachinesRequest *ListMachinesRequest) (*ListMachinesResponse, error) {
	return d.listMachines(listMachinesRequest)
}

// GetVolumeIDs parses volume names from pv specs
func (d *FakeDriver) GetVolumeIDs(getVolumeIDs *GetVolumeIDsRequest) (*GetVolumeIDsResponse, error) {
	return d.getVolumeIDs(getVolumeIDs)
}
