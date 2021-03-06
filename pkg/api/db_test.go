// Copyright (c) 2018 Huawei Technologies Co., Ltd. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/opensds/opensds/pkg/context"
	"github.com/opensds/opensds/pkg/db"
	"github.com/opensds/opensds/pkg/model"
	. "github.com/opensds/opensds/testutils/collection"
	dbtest "github.com/opensds/opensds/testutils/db/testing"
)

func TestCreateVolumeDBEntry(t *testing.T) {
	var in = &model.VolumeSpec{
		BaseModel:   &model.BaseModel{},
		Name:        "volume sample",
		Description: "This is a sample volume for testing",
		Size:        int64(1),
		Status:      model.VolumeCreating,
	}

	// Test case 1: Everything should work well.
	mockClient := new(dbtest.Client)
	mockClient.On("CreateVolume", context.NewAdminContext(), in).Return(&SampleVolumes[0], nil)
	db.C = mockClient

	var expected = &SampleVolumes[0]
	result, err := CreateVolumeDBEntry(context.NewAdminContext(), in)
	if err != nil {
		t.Errorf("Failed to create volume asynchronously, err is %v\n", err)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v\n", expected, result)
	}

	// Test case 2: The size of volume created should be greater than zero.
	in.Size = int64(-2)
	mockClient = new(dbtest.Client)
	mockClient.On("CreateVolume", context.NewAdminContext(), in).Return(&SampleVolumes[0], nil)
	db.C = mockClient
	_, err = CreateVolumeDBEntry(context.NewAdminContext(), in)
	expectedError := fmt.Sprintf("invalid volume size: %d", in.Size)
	if err == nil {
		t.Errorf("Expected Non-%v, got %v\n", nil, err)
	} else {
		if expectedError != err.Error() {
			t.Errorf("Expected Non-%v, got %v\n", expectedError, err.Error())
		}
	}
}

func TestCreateVolumeFromSnapshotDBEntry(t *testing.T) {
	var in = &model.VolumeSpec{
		BaseModel:   &model.BaseModel{},
		Name:        "volume sample",
		Description: "This is a sample volume for testing",
		Size:        int64(1),
		Status:      model.VolumeCreating,
		SnapshotId:  "3769855c-a102-11e7-b772-17b880d2f537",
	}
	var snap = &model.VolumeSnapshotSpec{
		BaseModel: &model.BaseModel{
			Id: "3769855c-a102-11e7-b772-17b880d2f537",
		},
		Size:   int64(1),
		Status: model.VolumeSnapAvailable,
	}

	// Test case 1: Everything should work well.
	mockClient := new(dbtest.Client)
	mockClient.On("CreateVolume", context.NewAdminContext(), in).Return(&SampleVolumes[1], nil)
	mockClient.On("GetVolumeSnapshot", context.NewAdminContext(), "3769855c-a102-11e7-b772-17b880d2f537").Return(snap, nil)
	db.C = mockClient

	var expected = &SampleVolumes[1]
	result, err := CreateVolumeDBEntry(context.NewAdminContext(), in)
	if err != nil {
		t.Errorf("Failed to create volume with snapshot, err is %v\n", err)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v\n", expected, result)
	}

	// Test case 2: The status of volume snapshot should always be available.
	snap.Status = model.VolumeSnapError
	mockClient = new(dbtest.Client)
	mockClient.On("CreateVolume", context.NewAdminContext(), in).Return(&SampleVolumes[1], nil)
	mockClient.On("GetVolumeSnapshot", context.NewAdminContext(), "3769855c-a102-11e7-b772-17b880d2f537").Return(snap, nil)
	db.C = mockClient

	_, err = CreateVolumeDBEntry(context.NewAdminContext(), in)
	expectedError := "only if the snapshot is available, the volume can be created"
	if err == nil {
		t.Errorf("Expected Non-%v, got %v\n", nil, err)
	} else {
		if expectedError != err.Error() {
			t.Errorf("Expected Non-%v, got %v\n", expectedError, err.Error())
		}
	}

	// Test case 3: Size of volume should always be equal to or bigger than
	// size of the snapshot.
	snap.Status, snap.Size = model.VolumeSnapAvailable, 10
	mockClient = new(dbtest.Client)
	mockClient.On("CreateVolume", context.NewAdminContext(), in).Return(&SampleVolumes[1], nil)
	mockClient.On("GetVolumeSnapshot", context.NewAdminContext(), "3769855c-a102-11e7-b772-17b880d2f537").Return(snap, nil)
	db.C = mockClient

	_, err = CreateVolumeDBEntry(context.NewAdminContext(), in)
	expectedError = "size of volume must be equal to or bigger than size of the snapshot"
	if err == nil {
		t.Errorf("Expected Non-%v, got %v\n", nil, err)
	} else {
		if expectedError != err.Error() {
			t.Errorf("Expected Non-%v, got %v\n", expectedError, err.Error())
		}
	}
}

func TestDeleteVolumeDBEntry(t *testing.T) {
	var vol = &model.VolumeSpec{
		BaseModel: &model.BaseModel{
			Id: "bd5b12a8-a101-11e7-941e-d77981b584d8",
		},
		Status:    model.VolumeAvailable,
		ProfileId: "3769855c-a102-11e7-b772-17b880d2f537",
		PoolId:    "3762355c-a102-11e7-b772-17b880d2f537",
	}
	var in = &model.VolumeSpec{
		BaseModel: &model.BaseModel{
			Id: "bd5b12a8-a101-11e7-941e-d77981b584d8",
		},
		Status:    model.VolumeDeleting,
		ProfileId: "3769855c-a102-11e7-b772-17b880d2f537",
		PoolId:    "3762355c-a102-11e7-b772-17b880d2f537",
	}

	// Test case 1: Everything should work well.
	mockClient := new(dbtest.Client)
	mockClient.On("DeleteVolume", context.NewAdminContext(), vol.Id).Return(nil)
	mockClient.On("ListSnapshotsByVolumeId", context.NewAdminContext(), vol.Id).Return(nil, nil)
	mockClient.On("ListAttachmentsByVolumeId", context.NewAdminContext(), vol.Id).Return(nil, nil)
	mockClient.On("UpdateVolume", context.NewAdminContext(), in).Return(nil, nil)
	db.C = mockClient

	err := DeleteVolumeDBEntry(context.NewAdminContext(), vol)
	if err != nil {
		t.Errorf("Failed to delete volume, err is %v\n", err)
	}

	// Test case 2: Volume to be deleted should not contain any snapshots.
	var sampleSnapshots = []*model.VolumeSnapshotSpec{&SampleSnapshots[0]}
	// Considering vol has been updated inisde DeleteVolumeDBEntry, so the status
	// should be rolled back here.
	vol.Status = model.VolumeAvailable
	mockClient = new(dbtest.Client)
	mockClient.On("DeleteVolume", context.NewAdminContext(), vol.Id).Return(nil)
	mockClient.On("ListSnapshotsByVolumeId", context.NewAdminContext(), vol.Id).Return(sampleSnapshots, nil)
	mockClient.On("ListAttachmentsByVolumeId", context.NewAdminContext(), vol.Id).Return(nil, nil)
	mockClient.On("UpdateVolume", context.NewAdminContext(), in).Return(nil, nil)
	db.C = mockClient

	err = DeleteVolumeDBEntry(context.NewAdminContext(), vol)
	expectedError := fmt.Sprintf("volume %s can not be deleted, because it still has snapshots", in.Id)
	if err == nil {
		t.Errorf("Expected Non-%v, got %v\n", nil, err)
	} else {
		if expectedError != err.Error() {
			t.Errorf("Expected Non-%v, got %v\n", expectedError, err.Error())
		}
	}

	// Test case 3: Volume to be deleted should not be in-use.
	var sampleAttachments = []*model.VolumeAttachmentSpec{&SampleAttachments[0]}
	// Considering vol has been updated inisde DeleteVolumeDBEntry, so the status
	// should be rolled back here.
	vol.Status = model.VolumeAvailable
	mockClient = new(dbtest.Client)
	mockClient.On("DeleteVolume", context.NewAdminContext(), vol.Id).Return(nil)
	mockClient.On("ListSnapshotsByVolumeId", context.NewAdminContext(), vol.Id).Return(nil, nil)
	mockClient.On("ListAttachmentsByVolumeId", context.NewAdminContext(), vol.Id).Return(sampleAttachments, nil)
	mockClient.On("UpdateVolume", context.NewAdminContext(), in).Return(nil, nil)
	db.C = mockClient

	err = DeleteVolumeDBEntry(context.NewAdminContext(), vol)
	expectedError = fmt.Sprintf("volume %s can not be deleted, because it's in use", in.Id)
	if err == nil {
		t.Errorf("Expected Non-%v, got %v\n", nil, err)
	} else {
		if expectedError != err.Error() {
			t.Errorf("Expected Non-%v, got %v\n", expectedError, err.Error())
		}
	}
}

func TestExtendVolumeDBEntry(t *testing.T) {
	var vol = &model.VolumeSpec{
		BaseModel: &model.BaseModel{
			Id: "bd5b12a8-a101-11e7-941e-d77981b584d8",
		},
		Status: model.VolumeAvailable,
		Size:   2,
	}
	var in = &model.VolumeSpec{
		BaseModel: &model.BaseModel{
			Id: "bd5b12a8-a101-11e7-941e-d77981b584d8",
		},
		Status: model.VolumeExtending,
		Size:   2,
	}

	// Test case 1: Everything should work well.
	mockClient := new(dbtest.Client)
	mockClient.On("GetVolume", context.NewAdminContext(), "bd5b12a8-a101-11e7-941e-d77981b584d8").Return(vol, nil)
	mockClient.On("ExtendVolume", context.NewAdminContext(), in).Return(nil, nil)
	db.C = mockClient
	_, err := ExtendVolumeDBEntry(context.NewAdminContext(), vol.Id, &model.ExtendVolumeSpec{NewSize: 20})
	if err != nil {
		t.Errorf("Failed to extend volume: %v\n", err)
	}

	// Test case 2: The status of volume should always be available.
	vol.Status = model.VolumeCreating
	mockClient = new(dbtest.Client)
	mockClient.On("GetVolume", context.NewAdminContext(), "bd5b12a8-a101-11e7-941e-d77981b584d8").Return(vol, nil)
	mockClient.On("ExtendVolume", context.NewAdminContext(), in).Return(nil, nil)
	db.C = mockClient
	_, err = ExtendVolumeDBEntry(context.NewAdminContext(), vol.Id, &model.ExtendVolumeSpec{NewSize: 20})
	expectedError := "the status of the volume to be extended must be available!"
	if err == nil {
		t.Errorf("Expected Non-%v, got %v\n", nil, err)
	} else {
		if expectedError != err.Error() {
			t.Errorf("Expected Non-%v, got %v\n", expectedError, err.Error())
		}
	}

	// Test case 3: The extended size should always be larger than current size.
	vol.Size, vol.Status = 20, model.VolumeAvailable
	in.Size = 20
	mockClient = new(dbtest.Client)
	mockClient.On("GetVolume", context.NewAdminContext(), "bd5b12a8-a101-11e7-941e-d77981b584d8").Return(vol, nil)
	mockClient.On("ExtendVolume", context.NewAdminContext(), in).Return(nil, nil)
	db.C = mockClient
	_, err = ExtendVolumeDBEntry(context.NewAdminContext(), vol.Id, &model.ExtendVolumeSpec{NewSize: 2})
	expectedError = "new size for extend must be greater than current size." +
		"(current: 20 GB, extended: 2 GB)."
	if err == nil {
		t.Errorf("Expected Non-%v, got %v\n", nil, err)
	} else {
		if expectedError != err.Error() {
			t.Errorf("Expected Non-%v, got %v\n", expectedError, err.Error())
		}
	}
}

func TestCreateVolumeAttachmentDBEntry(t *testing.T) {
	var m = map[string]string{"a": "a"}

	var req = &model.VolumeAttachmentSpec{
		BaseModel: &model.BaseModel{},
		VolumeId:  "bd5b12a8-a101-11e7-941e-d77981b584d8",
		Metadata:  m,
		Status:    "creating",
	}
	var vol = &model.VolumeSpec{
		BaseModel: &model.BaseModel{
			Id: "bd5b12a8-a101-11e7-941e-d77981b584d8",
		},
		Status: "available",
	}
	mockClient := new(dbtest.Client)
	mockClient.On("GetVolume", context.NewAdminContext(), "bd5b12a8-a101-11e7-941e-d77981b584d8").Return(vol, nil)
	mockClient.On("CreateVolumeAttachment", context.NewAdminContext(), req).Return(&SampleAttachments[0], nil)
	db.C = mockClient

	var expected = &SampleAttachments[0]

	result, err := CreateVolumeAttachmentDBEntry(context.NewAdminContext(), req)

	if err != nil {
		t.Errorf("Failed to create volume attachment, err is %v\n", err)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v\n", expected, result)
	}
}

func TestCreateVolumeSnapshotDBEntry(t *testing.T) {
	var m = map[string]string{"a": "a"}
	var vol = &model.VolumeSpec{
		BaseModel: &model.BaseModel{
			Id: "bd5b12a8-a101-11e7-941e-d77981b584d8",
		},
		Size:   1,
		Status: "available",
	}
	var req = &model.VolumeSnapshotSpec{
		BaseModel:   &model.BaseModel{},
		VolumeId:    "bd5b12a8-a101-11e7-941e-d77981b584d8",
		Name:        "sample-snapshot-01",
		Description: "This is the first sample snapshot for testing",
		Size:        int64(1),
		Status:      "creating",
		Metadata:    m,
	}

	mockClient := new(dbtest.Client)
	mockClient.On("GetVolume", context.NewAdminContext(), "bd5b12a8-a101-11e7-941e-d77981b584d8").Return(vol, nil)
	mockClient.On("CreateVolumeSnapshot", context.NewAdminContext(), req).Return(&SampleSnapshots[0], nil)
	db.C = mockClient

	var expected = &SampleSnapshots[0]
	result, err := CreateVolumeSnapshotDBEntry(context.NewAdminContext(), req)

	if err != nil {
		t.Errorf("Failed to create volume snapshot, err is %v\n", err)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v\n", expected, result)
	}
}

func TestDeleteVolumeSnapshotDBEntry(t *testing.T) {
	var req = &model.VolumeSnapshotSpec{
		BaseModel: &model.BaseModel{
			Id: "3769855c-a102-11e7-b772-17b880d2f537",
		},
		VolumeId: "bd5b12a8-a101-11e7-941e-d77981b584d8",
		Status:   "available",
	}

	mockClient := new(dbtest.Client)
	mockClient.On("UpdateVolumeSnapshot", context.NewAdminContext(), "3769855c-a102-11e7-b772-17b880d2f537", req).Return(nil, nil)
	mockClient.On("GetVolume", context.NewAdminContext(), req.VolumeId).Return(nil, nil)
	db.C = mockClient

	err := DeleteVolumeSnapshotDBEntry(context.NewAdminContext(), req)

	if err != nil {
		t.Errorf("Failed to delete volume snapshot, err is %v\n", err)
	}
}
