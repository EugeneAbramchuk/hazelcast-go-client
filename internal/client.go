// Copyright (c) 2008-2018, Hazelcast, Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License")
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package internal

import (
	"github.com/hazelcast/hazelcast-go-client/config"
	"github.com/hazelcast/hazelcast-go-client/core"
	"github.com/hazelcast/hazelcast-go-client/internal/common"
	"github.com/hazelcast/hazelcast-go-client/internal/serialization"
)

type HazelcastClient struct {
	ClientConfig         *config.ClientConfig
	InvocationService    *invocationService
	PartitionService     *partitionService
	SerializationService *serialization.SerializationService
	LifecycleService     *lifecycleService
	ConnectionManager    *connectionManager
	ListenerService      *listenerService
	ClusterService       *clusterService
	ProxyManager         *proxyManager
	LoadBalancer         *randomLoadBalancer
	HeartBeatService     *heartBeatService
}

func NewHazelcastClient(config *config.ClientConfig) (*HazelcastClient, error) {
	client := HazelcastClient{ClientConfig: config}
	err := client.init()
	return &client, err
}

func (c *HazelcastClient) GetMap(name string) (core.IMap, error) {
	mp, err := c.GetDistributedObject(common.ServiceNameMap, name)
	if err != nil {
		return nil, err
	}
	return mp.(core.IMap), nil
}

func (c *HazelcastClient) GetList(name string) (core.IList, error) {
	list, err := c.GetDistributedObject(common.ServiceNameList, name)
	if err != nil {
		return nil, err
	}
	return list.(core.IList), nil
}

func (c *HazelcastClient) GetSet(name string) (core.ISet, error) {
	set, err := c.GetDistributedObject(common.ServiceNameSet, name)
	if err != nil {
		return nil, err
	}
	return set.(core.ISet), nil
}
func (c *HazelcastClient) GetReplicatedMap(name string) (core.ReplicatedMap, error) {
	mp, err := c.GetDistributedObject(common.ServiceNameReplicatedMap, name)
	if err != nil {
		return nil, err
	}
	return mp.(core.ReplicatedMap), err
}

func (c *HazelcastClient) GetMultiMap(name string) (core.MultiMap, error) {
	mmp, err := c.GetDistributedObject(common.ServiceNameMultiMap, name)
	if err != nil {
		return nil, err
	}
	return mmp.(core.MultiMap), err
}

func (c *HazelcastClient) GetFlakeIdGenerator(name string) (core.FlakeIdGenerator, error) {
	flakeIdGenerator, err := c.GetDistributedObject(common.ServiceNameIdGenerator, name)
	if err != nil {
		return nil, err
	}
	return flakeIdGenerator.(core.FlakeIdGenerator), err
}

func (c *HazelcastClient) GetTopic(name string) (core.ITopic, error) {
	topic, err := c.GetDistributedObject(common.ServiceNameTopic, name)
	if err != nil {
		return nil, err
	}
	return topic.(core.ITopic), nil
}

func (c *HazelcastClient) GetQueue(name string) (core.IQueue, error) {
	queue, err := c.GetDistributedObject(common.ServiceNameQueue, name)
	if err != nil {
		return nil, err
	}
	return queue.(core.IQueue), nil
}

func (c *HazelcastClient) GetRingbuffer(name string) (core.Ringbuffer, error) {
	rb, err := c.GetDistributedObject(common.ServiceNameRingbufferService, name)
	if err != nil {
		return nil, err
	}
	return rb.(core.Ringbuffer), nil
}

func (c *HazelcastClient) GetPNCounter(name string) (core.PNCounter, error) {
	counter, err := c.GetDistributedObject(common.ServiceNamePNCounter, name)
	if err != nil {
		return nil, err
	}
	return counter.(core.PNCounter), nil
}

func (c *HazelcastClient) GetDistributedObject(serviceName string, name string) (core.IDistributedObject, error) {
	var clientProxy, err = c.ProxyManager.getOrCreateProxy(serviceName, name)
	if err != nil {
		return nil, err
	}
	return clientProxy, nil
}

func (c *HazelcastClient) GetCluster() core.ICluster {
	return c.ClusterService
}

func (c *HazelcastClient) GetLifecycle() core.ILifecycle {
	return c.LifecycleService
}

func (c *HazelcastClient) init() error {
	c.LifecycleService = newLifecycleService(c.ClientConfig)
	c.ConnectionManager = newConnectionManager(c)
	c.HeartBeatService = newHeartBeatService(c)
	c.InvocationService = newInvocationService(c)
	c.ClusterService = newClusterService(c, c.ClientConfig)
	c.ListenerService = newListenerService(c)
	c.PartitionService = newPartitionService(c)
	c.ProxyManager = newProxyManager(c)
	c.LoadBalancer = newRandomLoadBalancer(c.ClusterService)
	c.SerializationService = serialization.NewSerializationService(c.ClientConfig.SerializationConfig())
	err := c.ClusterService.start()
	if err != nil {
		return err
	}
	c.HeartBeatService.start()
	c.PartitionService.start()
	c.LifecycleService.fireLifecycleEvent(LifecycleStateStarted)
	return nil
}

func (c *HazelcastClient) Shutdown() {
	if c.LifecycleService.isLive.Load().(bool) {
		c.LifecycleService.fireLifecycleEvent(LifecycleStateShuttingDown)
		c.PartitionService.shutdown()
		c.InvocationService.shutdown()
		c.HeartBeatService.shutdown()
		c.LifecycleService.fireLifecycleEvent(LifecycleStateShutdown)
	}
}
