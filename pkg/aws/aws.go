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

// Package aws contains the cloud provider specific implementations to manage machines
package aws

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/gardener/machine-controller-aws/pkg/spi"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/driver"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/codes"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/status"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog"
)

// Driver is the driver struct for holding AWS machine information
type Driver struct {
	SPI spi.SessionProviderInterface
}

const (
	resourceTypeInstance = "instance"
	resourceTypeVolume   = "volume"
)

// NewAWSDriver returns an empty AWSDriver object
func NewAWSDriver(spi spi.SessionProviderInterface) driver.Driver {
	return &Driver{
		SPI: spi,
	}
}

// CreateMachine returns a newly created fake driver
func (d *Driver) CreateMachine(ctx context.Context, req *driver.CreateMachineRequest) (*driver.CreateMachineResponse, error) {
	var (
		exists       bool
		userData     []byte
		machine      = req.Machine
		secret       = req.Secret
		machineClass = req.MachineClass
	)

	// Log messages to track request
	klog.V(2).Infof("Machine creation request has been recieved for %q", req.Machine.Name)

	providerSpec, err := decodeProviderSpecAndSecret(machineClass, secret)
	if err != nil {
		return nil, err
	}

	svc, err := d.createSVC(secret, providerSpec.Region)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if userData, exists = secret.Data["userData"]; !exists {
		return nil, status.Error(codes.Internal, "userData doesn't exist")
	}
	UserDataEnc := base64.StdEncoding.EncodeToString([]byte(userData))

	var imageIds []*string
	imageID := aws.String(providerSpec.AMI)
	imageIds = append(imageIds, imageID)

	describeImagesRequest := ec2.DescribeImagesInput{
		ImageIds: imageIds,
	}
	output, err := svc.DescribeImages(&describeImagesRequest)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var blkDeviceMappings []*ec2.BlockDeviceMapping
	deviceName := output.Images[0].RootDeviceName
	volumeSize := providerSpec.BlockDevices[0].Ebs.VolumeSize
	volumeType := providerSpec.BlockDevices[0].Ebs.VolumeType
	blkDeviceMapping := ec2.BlockDeviceMapping{
		DeviceName: deviceName,
		Ebs: &ec2.EbsBlockDevice{
			VolumeSize: &volumeSize,
			VolumeType: &volumeType,
		},
	}
	if volumeType == "io1" {
		blkDeviceMapping.Ebs.Iops = &providerSpec.BlockDevices[0].Ebs.Iops
	}
	blkDeviceMappings = append(blkDeviceMappings, &blkDeviceMapping)

	// Add tags to the created machine
	tagList := []*ec2.Tag{}
	for idx, element := range providerSpec.Tags {
		if idx == "Name" {
			// Name tag cannot be set, as its used to identify backing machine object
			klog.Warning("Name tag cannot be set on AWS instance, as its used to identify backing machine object")
			continue
		}
		newTag := ec2.Tag{
			Key:   aws.String(idx),
			Value: aws.String(element),
		}
		tagList = append(tagList, &newTag)
	}
	nameTag := ec2.Tag{
		Key:   aws.String("Name"),
		Value: aws.String(machine.Name),
	}
	tagList = append(tagList, &nameTag)

	tagInstance := &ec2.TagSpecification{
		ResourceType: aws.String("instance"),
		Tags:         tagList,
	}

	// Specify the details of the machine that you want to create.
	inputConfig := ec2.RunInstancesInput{
		// An Amazon Linux AMI ID for t2.micro machines in the us-west-2 region
		ImageId:      aws.String(providerSpec.AMI),
		InstanceType: aws.String(providerSpec.MachineType),
		MinCount:     aws.Int64(1),
		MaxCount:     aws.Int64(1),
		UserData:     &UserDataEnc,
		KeyName:      aws.String(providerSpec.KeyName),
		SubnetId:     aws.String(providerSpec.NetworkInterfaces[0].SubnetID),
		IamInstanceProfile: &ec2.IamInstanceProfileSpecification{
			Name: &(providerSpec.IAM.Name),
		},
		SecurityGroupIds:    []*string{aws.String(providerSpec.NetworkInterfaces[0].SecurityGroupIDs[0])},
		BlockDeviceMappings: blkDeviceMappings,
		TagSpecifications:   []*ec2.TagSpecification{tagInstance},
	}

	runResult, err := svc.RunInstances(&inputConfig)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	response := &driver.CreateMachineResponse{
		ProviderID: encodeProviderID(providerSpec.Region, *runResult.Instances[0].InstanceId),
		NodeName:   *runResult.Instances[0].PrivateDnsName,
	}

	klog.V(2).Infof("VM with Provider-ID: %q created for Machine: %q", response.ProviderID, machine.Name)
	return response, nil
}

// DeleteMachine deletes the fake machine
func (d *Driver) DeleteMachine(ctx context.Context, req *driver.DeleteMachineRequest) (*driver.DeleteMachineResponse, error) {
	var (
		secret       = req.Secret
		machineClass = req.MachineClass
	)

	// Log messages to track delete request
	klog.V(2).Infof("Machine deletion request has been recieved for %q", req.Machine.Name)
	defer klog.V(2).Infof("Machine deletion request has been processed for %q", req.Machine.Name)

	providerSpec, err := decodeProviderSpecAndSecret(machineClass, secret)
	if err != nil {
		return nil, err
	}

	instances, err := d.getInstancesFromMachineName(req.Machine.Name, providerSpec, secret)
	if err != nil {
		return nil, err
	}

	svc, err := d.createSVC(secret, providerSpec.Region)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	for _, instance := range instances {
		input := &ec2.TerminateInstancesInput{
			InstanceIds: []*string{
				aws.String(*instance.InstanceId),
			},
			DryRun: aws.Bool(false),
		}
		_, err = svc.TerminateInstances(input)
		if err != nil {
			klog.V(2).Infof("VM %q for Machine %q couldn't be terminated: %s",
				*instance.InstanceId,
				req.Machine.Name,
				err.Error(),
			)
			return nil, status.Error(codes.Internal, err.Error())
		}

		klog.V(2).Infof("VM %q for Machine %q was terminated succesfully", *instance.InstanceId, req.Machine.Name)
	}

	return &driver.DeleteMachineResponse{}, nil
}

// GetMachineStatus returns the machine status
func (d *Driver) GetMachineStatus(ctx context.Context, req *driver.GetMachineStatusRequest) (*driver.GetMachineStatusResponse, error) {
	var (
		secret       = req.Secret
		machineClass = req.MachineClass
	)

	// Log messages to track start and end of request
	klog.V(2).Infof("Get request has been recieved for %q", req.Machine.Name)

	providerSpec, err := decodeProviderSpecAndSecret(machineClass, secret)
	if err != nil {
		return nil, err
	}

	instances, err := d.getInstancesFromMachineName(req.Machine.Name, providerSpec, secret)
	if err != nil {
		return nil, err
	} else if len(instances) > 1 {
		instanceIDs := []string{}
		for _, instance := range instances {
			instanceIDs = append(instanceIDs, *instance.InstanceId)
		}

		errMessage := fmt.Sprintf("AWS plugin is returning multiple VM instances backing this machine object. IDs for all backing VMs - %v ", instanceIDs)
		return nil, status.Error(codes.OutOfRange, errMessage)
	}

	requiredInstance := instances[0]

	response := &driver.GetMachineStatusResponse{
		NodeName:   *requiredInstance.PrivateDnsName,
		ProviderID: encodeProviderID(providerSpec.Region, *requiredInstance.InstanceId),
	}

	klog.V(2).Infof("Machine get request has been processed successfully for %q", req.Machine.Name)
	return response, nil
}

// ListMachines returns the list of VMs for a given machineClass
func (d *Driver) ListMachines(ctx context.Context, req *driver.ListMachinesRequest) (*driver.ListMachinesResponse, error) {
	var (
		machineClass = req.MachineClass
		secret       = req.Secret
	)

	// Log messages to track start and end of request
	klog.V(2).Infof("List machines request has been recieved for %q", machineClass.Name)

	providerSpec, err := decodeProviderSpecAndSecret(machineClass, secret)
	if err != nil {
		return nil, err
	}

	clusterName := ""
	nodeRole := ""

	for key := range providerSpec.Tags {
		if strings.Contains(key, "kubernetes.io/cluster/") {
			clusterName = key
		} else if strings.Contains(key, "kubernetes.io/role/") {
			nodeRole = key
		}
	}

	svc, err := d.createSVC(secret, providerSpec.Region)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	input := ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name: aws.String("tag-key"),
				Values: []*string{
					&clusterName,
				},
			},
			&ec2.Filter{
				Name: aws.String("tag-key"),
				Values: []*string{
					&nodeRole,
				},
			},
			&ec2.Filter{
				Name: aws.String("instance-state-name"),
				Values: []*string{
					aws.String("pending"),
					aws.String("running"),
					aws.String("stopping"),
					aws.String("stopped"),
				},
			},
		},
	}

	runResult, err := svc.DescribeInstances(&input)
	if err != nil {
		klog.Errorf("AWS plugin is returning error while describe instances request is sent: %s", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	listOfVMs := make(map[string]string)
	for _, reservation := range runResult.Reservations {
		for _, instance := range reservation.Instances {

			machineName := ""
			for _, tag := range instance.Tags {
				if *tag.Key == "Name" {
					machineName = *tag.Value
					break
				}
			}
			listOfVMs[encodeProviderID(providerSpec.Region, *instance.InstanceId)] = machineName
		}
	}

	klog.V(2).Infof("List machines request has been processed successfully")
	// Core logic ends here.
	resp := &driver.ListMachinesResponse{
		MachineList: listOfVMs,
	}
	return resp, nil
}

// GetVolumeIDs parses volume names from pv specs
func (d *Driver) GetVolumeIDs(ctx context.Context, req *driver.GetVolumeIDsRequest) (*driver.GetVolumeIDsResponse, error) {
	var (
		volumeIDs   []string
		volumeSpecs []*corev1.PersistentVolumeSpec
	)

	// Log messages to track start and end of request
	klog.V(2).Infof("GetVolumeIDs request has been recieved for %q", req.PVSpecs)

	for i := range req.PVSpecs {
		spec := volumeSpecs[i]
		if spec.AWSElasticBlockStore == nil {
			// Not an aws volume
			continue
		}
		volumeID := spec.AWSElasticBlockStore.VolumeID
		volumeIDs = append(volumeIDs, volumeID)
	}

	klog.V(2).Infof("GetVolumeIDs machines request has been processed successfully. \nList: %v", volumeIDs)

	resp := &driver.GetVolumeIDsResponse{
		VolumeIDs: volumeIDs,
	}
	return resp, nil
}
