/*
 * Copyright 2022 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package itf

import (
	"context"
	"io/fs"
	"time"
)

type ContainerEngineHandler interface {
	ListContainers(ctx context.Context, filter [][2]string) ([]Container, error)
	ContainerCreate(ctx context.Context, ctrConf Container) (id string, err error)
	ContainerInfo(ctx context.Context, id string) (Container, error)
	ImageInfo(ctx context.Context, id string) (Image, error)
}

type Image struct {
	ID      string   `json:"id"`
	Created string   `json:"created"`
	Size    int64    `json:"size"`
	Arch    string   `json:"arch"`
	Tags    []string `json:"tags"`
	Digests []string `json:"digests"`
}

type NetworkType string

type Network struct {
	ID   string      `json:"id"`
	Name string      `json:"name"`
	Type NetworkType `json:"type"`
}

type PortType string

type Port struct {
	Number   int      `json:"number"`
	Protocol PortType `json:"protocol"`
	Bindings []PortBinding
}

type PortBinding struct {
	Number    int    `json:"number"`
	Interface string `json:"interface"`
}

type MountType string

type Mount struct {
	Type     MountType         `json:"type"`
	Source   string            `json:"source"`
	Target   string            `json:"target"`
	ReadOnly bool              `json:"read_only"`
	Labels   map[string]string `json:"labels,omitempty"`
	Size     int64             `json:"size,omitempty"`
	Mode     fs.FileMode       `json:"mode,omitempty"`
}

type RestartStrategy string

type RunConfig struct {
	RestartStrategy RestartStrategy `json:"strategy"`
	Retries         int             `json:"retries"`
	RemoveAfterRun  bool            `json:"remove_after_run"`
	StopTimeout     *time.Duration  `json:"stop_timeout"`
}

type ContainerState string

type Container struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	State     ContainerState    `json:"state"`
	Created   string            `json:"created"`
	Started   string            `json:"started"`
	Image     string            `json:"image"`
	ImageID   string            `json:"image_id"`
	EnvVars   map[string]string `json:"env_vars"`
	Labels    map[string]string `json:"labels"`
	Mounts    []Mount           `json:"mounts"`
	Ports     []Port            `json:"ports"`
	Networks  []ContainerNet    `json:"networks"`
	RunConfig RunConfig         `json:"run_config"`
}

type ContainerNet struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	DomainNames []string `json:"domain_names"`
	Gateway     string   `json:"gateway"`
	IPAddress   string   `json:"ip_address"`
	MacAddress  string   `json:"mac_address"`
}