// Copyright 2021 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kuberbacproxy

import (
	"context"
	"errors"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	resourcesv1alpha1 "github.com/gardener/gardener/pkg/apis/resources/v1alpha1"
	"github.com/gardener/gardener/pkg/client/kubernetes"
	"github.com/gardener/gardener/pkg/component"
	"github.com/gardener/gardener/pkg/utils/managedresources"
)

const (
	managedResourceName = "shoot-node-logging"
)

// New creates a new instance of kubeRBACProxy for the kube-rbac-proxy.
// Deprecated: This component is deprecated and will be removed after gardener/gardener@v1.78 has been released.
// TODO(rfranzke): Delete the `shoot-node-logging` ManagedResource and drop this component after gardener/gardener@v1.78 has been released.
func New(client client.Client, namespace string) (component.Deployer, error) {
	if client == nil {
		return nil, errors.New("client cannot be nil")
	}

	if len(namespace) == 0 {
		return nil, errors.New("namespace cannot be empty")
	}

	return &kubeRBACProxy{client: client, namespace: namespace}, nil
}

type kubeRBACProxy struct {
	// client to create resources with.
	client client.Client
	// namespace in the seed cluster.
	namespace string
}

func (k *kubeRBACProxy) Deploy(ctx context.Context) error {
	var (
		kubeRBACProxyClusterRolebinding = &rbacv1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "gardener.cloud:logging:kube-rbac-proxy",
				Annotations: map[string]string{resourcesv1alpha1.Mode: resourcesv1alpha1.ModeIgnore},
			},
		}

		valitailClusterRole = &rbacv1.ClusterRole{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "gardener.cloud:logging:valitail",
				Annotations: map[string]string{resourcesv1alpha1.Mode: resourcesv1alpha1.ModeIgnore},
			},
		}

		valitailClusterRoleBinding = &rbacv1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "gardener.cloud:logging:valitail",
				Annotations: map[string]string{resourcesv1alpha1.Mode: resourcesv1alpha1.ModeIgnore},
			},
		}

		registry = managedresources.NewRegistry(kubernetes.ShootScheme, kubernetes.ShootCodec, kubernetes.ShootSerializer)
	)

	resources, err := registry.AddAllAndSerialize(
		kubeRBACProxyClusterRolebinding,
		valitailClusterRole,
		valitailClusterRoleBinding,
	)
	if err != nil {
		return err
	}

	return managedresources.CreateForShoot(ctx, k.client, k.namespace, managedResourceName, managedresources.LabelValueGardener, false, resources)
}

func (k *kubeRBACProxy) Destroy(ctx context.Context) error {
	return managedresources.DeleteForShoot(ctx, k.client, k.namespace, managedResourceName)
}
