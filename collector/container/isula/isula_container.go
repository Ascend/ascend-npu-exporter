package isula

type Config struct {
	Env []string `json:"Env,omitempty" platform:"linux"`
}

type DeviceInfo struct {
	PathInContainer string `json:"PathInContainer,omitempty" platform:"linux"`
}

type HostConfig struct {
	Devices    []DeviceInfo `json:"Devices,omitempty" platform:"linux"`
	Privileged bool         `json:"Privileged,omitempty" platform:"linux"`
}

type ContainerJson struct {
	Config     *Config     `json:"Config,omitempty" platform:"linux"`
	HostConfig *HostConfig `json:"HostConfig,omitempty" platform:"linux"`
}
