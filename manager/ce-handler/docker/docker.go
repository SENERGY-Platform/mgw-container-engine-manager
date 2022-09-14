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

package docker

import (
	"context"
	"deployment-manager/manager/ce-handler/itf"
	"deployment-manager/util"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

type Docker struct {
	client *client.Client
}

func New() (*Docker, error) {
	c, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &Docker{client: c}, nil
}

func (d *Docker) ListContainers(ctx context.Context, filter map[string]string) (map[string]*itf.Container, error) {
	var f filters.Args
	if filter != nil && len(filter) > 0 {
		f = filters.NewArgs()
		for k, v := range filter {
			f.Add(k, v)
		}
	}
	if cl, err := d.client.ContainerList(ctx, types.ContainerListOptions{Filters: f}); err != nil {
		return nil, err
	} else {
		cm := make(map[string]*itf.Container, len(cl))
		for _, c := range cl {
			container := &itf.Container{
				ID:       c.ID,
				ImageID:  c.ImageID,
				Image:    c.Image,
				Status:   c.State,
				Mounts:   parseContainerMounts(c.Mounts),
				Labels:   c.Labels,
				Networks: parseContainerNetworks(c.NetworkSettings.Networks),
			}
			if ci, err := d.client.ContainerInspect(ctx, c.ID); err != nil {
				util.Logger.Error(err)
			} else {
				container.Name = ci.Name
				container.StartPolicy = ci.HostConfig.RestartPolicy.Name
				container.CreatedAt = ci.Created
				container.StartedAt = ci.State.StartedAt
				container.Hostname = ci.Config.Hostname
				container.Env = parseContainerEnvVars(ci.Config.Env)
				container.Networks = parseContainerNetworks(ci.NetworkSettings.Networks)
				container.Ports = parseContainerPorts(ci.Config.ExposedPorts, ci.NetworkSettings.Ports)
			}
			cm[c.ID] = container
		}
		return cm, nil
	}
}

func (d *Docker) ContainerInfo(ctx context.Context, id string) (*itf.Container, error) {
	if c, err := d.client.ContainerInspect(ctx, id); err != nil {
		return nil, err
	} else {
		return &itf.Container{
			ID:          c.ID,
			Name:        c.Name,
			ImageID:     c.Image,
			Image:       c.Config.Image,
			Status:      c.State.Status,
			StartPolicy: c.HostConfig.RestartPolicy.Name,
			CreatedAt:   c.Created,
			StartedAt:   c.State.StartedAt,
			Hostname:    c.Config.Hostname,
			Env:         parseContainerEnvVars(c.Config.Env),
			Mounts:      parseContainerMounts(c.Mounts),
			Labels:      c.Config.Labels,
			Ports:       parseContainerPorts(c.Config.ExposedPorts, c.NetworkSettings.Ports),
			Networks:    parseContainerNetworks(c.NetworkSettings.Networks),
		}, nil
	}
}

func (d *Docker) ImageInfo(ctx context.Context, id string) (*itf.Image, error) {
	if i, _, err := d.client.ImageInspectWithRaw(ctx, id); err != nil {
		return nil, err
	} else {
		return &itf.Image{
			ID:      i.ID,
			Created: i.Created,
			Size:    i.Size,
			Arch:    i.Architecture,
			Tags:    i.RepoTags,
			Digests: i.RepoDigests,
		}, nil
	}
}