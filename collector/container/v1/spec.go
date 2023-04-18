/* Copyright(C) 2021-2023. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package v1 implement the containerd client
package v1

import "os"

// Spec is the base configuration for the container.
type Spec struct {
	// Linux is platform-specific configuration for Linux based containers.
	Linux *Linux `json:"linux,omitempty" platform:"linux"`
	// Version of the Open Container Initiative Runtime Specification with which the bundle complies.
	Version string `json:"ociVersion"`
	// Process for get capabilities
	Process *Process `json:"process,omitempty" platform:"linux"`
}

// Process is the base configuration for the container.
type Process struct {
	// Capabilities for container capabilities
	Capabilities *Capabilities `json:"capabilities,omitempty" platform:"linux"`
	// Env for container env
	Env []string `json:"env,omitempty" platform:"linux"`
}

// Capabilities is the base configuration for the container.
type Capabilities struct {
	// Permitted capabilities that can be used by container
	Permitted []string `json:"permitted,omitempty" platform:"linux"`
}

// Linux contains platform-specific configuration for Linux based containers.
type Linux struct {
	// CgroupsPath specifies the path to cgroups that are created and/or joined by the container.
	CgroupsPath string `json:"cgroupsPath,omitempty"`
	// Devices are a list of device nodes that are created for the container
	Devices []Device `json:"devices,omitempty"`
}

// Device linux device info
type Device struct {
	// UID of the device.
	UID *uint32 `json:"uid,omitempty"`
	// GID of the device.
	GID *uint32 `json:"gid,omitempty"`
	// FileMode permission bits for the device.
	FileMode *os.FileMode `json:"fileMode,omitempty"`
	// Path to the device.
	Path string `json:"mount_path"`
	// Type device type, block, char, etc.
	Type string `json:"type"`
	// Major is the device's major id.
	Major int64 `json:"major"`
	// Minor is the device's minor id
	Minor int64 `json:"minor"`
}
