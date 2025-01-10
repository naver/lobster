/*
 * Copyright (c) 2024-present NAVER Corp
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/golang/glog"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	conf    config
	timeout = time.Second
)

func init() {
	conf = setup()
	log.Println("k8s client configuration is loaded")
}

type Client struct {
	hostName string
	*kubernetes.Clientset
	timeout time.Duration
	cache   map[string]v1.Pod
}

func New() (Client, error) {
	if len(*conf.HostName) == 0 {
		return Client{}, errors.New("`client.hostName` is required")
	}
	config, err := rest.InClusterConfig()
	if err != nil {
		return Client{}, err
	}

	clientset, err := kubernetes.NewForConfig(config)

	return Client{*conf.HostName, clientset, timeout, map[string]v1.Pod{}}, err
}

func (c *Client) GetPods() map[string]v1.Pod {
	podMap := map[string]v1.Pod{}
	podList := v1.PodList{}

	ctx, cancel := context.WithTimeout(context.TODO(), c.timeout)
	defer cancel()

	data, err := c.RESTClient().
		Get().
		AbsPath(fmt.Sprintf("/api/v1/nodes/%s/proxy/pods", c.hostName)).
		DoRaw(ctx)
	if err != nil {
		glog.Warningf("using cached pod information: failed to make a request to the k8s API server or kubelet: %s", err.Error())
		return c.cache
	}

	if err := json.Unmarshal(data, &podList); err != nil {
		glog.Warningf("using cached pod information: failed to unmarshal the pod list response: %s", err.Error())
		return c.cache
	}

	for _, pod := range podList.Items {
		podMap[string(pod.UID)] = pod
	}

	c.cache = podMap

	return podMap
}
