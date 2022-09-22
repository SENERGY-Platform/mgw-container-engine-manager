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
	"deployment-manager/manager/handler/docker/util"
	"deployment-manager/manager/itf"
	dmUtil "deployment-manager/util"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"io"
	"strconv"
)

func (d *Docker) ListContainers(ctx context.Context, filter [][2]string) ([]itf.Container, error) {
	if cl, err := d.client.ContainerList(ctx, types.ContainerListOptions{All: true, Filters: util.GenFilterArgs(filter)}); err != nil {
		return nil, err
	} else {
		var csl []itf.Container
		for _, c := range cl {
			ctr := itf.Container{
				ID:       c.ID,
				State:    util.StateMap[c.State],
				ImageID:  c.ImageID,
				Labels:   c.Labels,
				Mounts:   util.ParseMountPoints(c.Mounts),
				Networks: util.ParseEndpointSettings(c.NetworkSettings.Networks),
			}
			if ci, err := d.client.ContainerInspect(ctx, c.ID); err != nil {
				dmUtil.Logger.Error(err)
			} else {
				ctr.Name = ci.Name
				if tc, err := util.ParseTimestamp(ci.Created); err != nil {
					dmUtil.Logger.Error(err)
				} else {
					ctr.Created = tc
				}
				if c.State == "running" {
					if ts, err := util.ParseTimestamp(ci.State.StartedAt); err != nil {
						dmUtil.Logger.Error(err)
					} else {
						ctr.Started = &ts
					}
				}
				ctr.Image = ci.Config.Image
				ctr.EnvVars = util.ParseEnv(ci.Config.Env)
				ctr.Ports = util.ParsePortSetAndMap(ci.Config.ExposedPorts, ci.NetworkSettings.Ports)
				if len(ci.HostConfig.Mounts) > 0 {
					ctr.Mounts = util.ParseMounts(ci.HostConfig.Mounts)
				}
				ctr.Networks = util.ParseEndpointSettings(ci.NetworkSettings.Networks)
				ctr.RunConfig = itf.RunConfig{
					RestartStrategy: util.RestartPolicyMap[ci.HostConfig.RestartPolicy.Name],
					Retries:         ci.HostConfig.RestartPolicy.MaximumRetryCount,
					RemoveAfterRun:  ci.HostConfig.AutoRemove,
					StopTimeout:     util.ParseStopTimeout(ci.Config.StopTimeout),
				}
			}
			csl = append(csl, ctr)
		}
		return csl, nil
	}
}

func (d *Docker) ContainerInfo(ctx context.Context, id string) (itf.Container, error) {
	var ctr itf.Container
	if c, err := d.client.ContainerInspect(ctx, id); err != nil {
		return ctr, err
	} else {
		var mts []itf.Mount
		if len(c.HostConfig.Mounts) > 0 {
			mts = util.ParseMounts(c.HostConfig.Mounts)
		} else {
			mts = util.ParseMountPoints(c.Mounts)
		}
		ctr = itf.Container{
			ID:       c.ID,
			Name:     c.Name,
			State:    util.StateMap[c.State.Status],
			Image:    c.Config.Image,
			ImageID:  c.Image,
			EnvVars:  util.ParseEnv(c.Config.Env),
			Labels:   c.Config.Labels,
			Mounts:   mts,
			Ports:    util.ParsePortSetAndMap(c.Config.ExposedPorts, c.NetworkSettings.Ports),
			Networks: util.ParseEndpointSettings(c.NetworkSettings.Networks),
			RunConfig: itf.RunConfig{
				RestartStrategy: util.RestartPolicyMap[c.HostConfig.RestartPolicy.Name],
				Retries:         c.HostConfig.RestartPolicy.MaximumRetryCount,
				RemoveAfterRun:  c.HostConfig.AutoRemove,
				StopTimeout:     util.ParseStopTimeout(c.Config.StopTimeout),
			},
		}
		if tc, err := util.ParseTimestamp(c.Created); err != nil {
			dmUtil.Logger.Error(err)
		} else {
			ctr.Created = tc
		}
		if c.State.Status == "running" {
			if ts, err := util.ParseTimestamp(c.State.StartedAt); err != nil {
				dmUtil.Logger.Error(err)
			} else {
				ctr.Started = &ts
			}
		}
	}
	return ctr, nil
}

func (d *Docker) ContainerCreate(ctx context.Context, ctrConf itf.Container) (string, error) {
	cConfig := &container.Config{
		AttachStdout: true,
		AttachStderr: true,
		Env:          util.GenEnv(ctrConf.EnvVars),
		Image:        ctrConf.Image,
		Labels:       ctrConf.Labels,
		StopTimeout:  util.GenStopTimeout(ctrConf.RunConfig.StopTimeout),
	}
	bindings, err := util.GenPortMap(ctrConf.Ports)
	if err != nil {
		return "", err
	}
	mts, err := util.GenMounts(ctrConf.Mounts)
	if err != nil {
		return "", err
	}
	hConfig := &container.HostConfig{
		PortBindings: bindings,
		RestartPolicy: container.RestartPolicy{
			Name:              util.RestartPolicyRMap[ctrConf.RunConfig.RestartStrategy],
			MaximumRetryCount: ctrConf.RunConfig.Retries,
		},
		AutoRemove: ctrConf.RunConfig.RemoveAfterRun,
		Mounts:     mts,
	}
	err = util.CheckNetworks(ctrConf.Networks)
	if err != nil {
		return "", err
	}
	var nConfig *network.NetworkingConfig
	if len(ctrConf.Networks) > 0 {
		nConfig = &network.NetworkingConfig{EndpointsConfig: map[string]*network.EndpointSettings{
			ctrConf.Networks[0].Name: {
				Aliases: ctrConf.Networks[0].DomainNames,
			},
		}}
	}
	res, err := d.client.ContainerCreate(ctx, cConfig, hConfig, nConfig, nil, ctrConf.Name)
	if err != nil {
		return "", err
	}
	if res.Warnings != nil && len(res.Warnings) > 0 {
		dmUtil.Logger.Warning(res.Warnings)
	}
	if len(ctrConf.Networks) > 1 {
		for i := 1; i < len(ctrConf.Networks); i++ {
			err := d.client.NetworkConnect(ctx, ctrConf.Networks[i].Name, res.ID, &network.EndpointSettings{
				Aliases: ctrConf.Networks[i].DomainNames,
			})
			if err != nil {
				err2 := d.ContainerRemove(ctx, res.ID)
				if err2 != nil {
					dmUtil.Logger.Error(err2)
				}
				return "", err
			}
		}
	}
	return res.ID, nil
}

func (d *Docker) ContainerRemove(ctx context.Context, id string) error {
	return d.client.ContainerRemove(ctx, id, types.ContainerRemoveOptions{
		RemoveVolumes: true,
	})
}

func (d *Docker) ContainerStart(ctx context.Context, id string) error {
	return d.client.ContainerStart(ctx, id, types.ContainerStartOptions{})
}

func (d *Docker) ContainerStop(ctx context.Context, id string) error {
	return d.client.ContainerStop(ctx, id, nil)
}

func (d *Docker) ContainerRestart(ctx context.Context, id string) error {
	return d.client.ContainerRestart(ctx, id, nil)
}

func (d *Docker) ContainerLog(ctx context.Context, id string, logOpt itf.LogOptions) (io.ReadCloser, error) {
	var lr util.LogReader
	clo := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	}
	if logOpt.Since != nil {
		clo.Since = util.GenTimestamp(*logOpt.Since)
	}
	if logOpt.Until != nil {
		clo.Until = util.GenTimestamp(*logOpt.Until)
	}
	if logOpt.MaxLines > 0 {
		clo.Tail = strconv.FormatInt(int64(logOpt.MaxLines), 10)
	}
	rc, err := d.client.ContainerLogs(ctx, id, clo)
	if err != nil {
		return &lr, err
	}
	return util.NewLogReader(rc), nil
}