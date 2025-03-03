// Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
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

package botanist

import (
	"context"
	"crypto/rand"
	"time"

	"github.com/gardener/gardener/pkg/utils"

	"github.com/gardener/gardener/pkg/operation/common"
	encryptionconfiguration "github.com/gardener/gardener/pkg/operation/etcdencryption"
	kutil "github.com/gardener/gardener/pkg/utils/kubernetes"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	apiserverconfigv1 "k8s.io/apiserver/pkg/apis/config/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// ApplyEncryptionConfiguration creates or updates a secret on the Seed
// which contains the encryption configuration that is necessary to encrypt the
// Kubernetes secrets in etcd.
//
// To mitigate data loss to a certain degree, the secret is also synced to the Garden cluster.
func (b *Botanist) ApplyEncryptionConfiguration(ctx context.Context) error {
	conf, err := b.createOrUpdateEncryptionConfiguration(ctx)
	if err != nil {
		return err
	}

	return b.syncEncryptionConfigurationToGarden(ctx, conf)
}

func (b *Botanist) createOrUpdateEncryptionConfiguration(ctx context.Context) (*apiserverconfigv1.EncryptionConfiguration, error) {
	var (
		secret = &corev1.Secret{ObjectMeta: kutil.ObjectMeta(b.Shoot.SeedNamespace, common.EtcdEncryptionSecretName)}
		conf   *apiserverconfigv1.EncryptionConfiguration
	)

	_, err := controllerutil.CreateOrUpdate(ctx, b.K8sSeedClient.Client(), secret, func() error {
		var err error
		conf, err = encryptionconfiguration.ReadSecret(secret)
		if err != nil {
			if !encryptionconfiguration.IsConfigurationNotFoundError(err) {
				return err
			}

			b.Logger.Info("Creating new etcd encryption configuration for Shoot")
			conf, err = encryptionconfiguration.NewPassiveConfiguration(time.Now(), rand.Reader)
			if err != nil {
				return err
			}
		}

		// When firstly created, the encryption configuration secret does not have a checksum annotation yet. This annotation will
		// only be added after all shoot secrets have been rewritten. In order to allow a smooth transition from un-encrypted to encrypted
		// etcd data we first make the configuration inactive, i.e., put the `identity` provider as first list in the entry. In the next
		// reconciliation we will detect that the annotation is set and then we can make it active, i.e., moving the `identity` provider
		// to the second list item. The reason for this is that new API servers would otherwise start with an active configuration and would
		// try to decrypt secrets in the etcd store (which would fail because they are not yet encrypted). Be aware that this will only happen
		// once during the first introduction of the encryption configuration.
		firstCreationOfEncryptionConfiguration := !metav1.HasAnnotation(secret.ObjectMeta, common.EtcdEncryptionChecksumAnnotationName)

		// We allow to force the API servers to not encrypt the secrets in etcd store. This is possible by annotating the etcd-encryption-secret
		// with 'shoot.gardener.cloud/etcd-encryption-force-plaintext-secrets=true'.
		forcePlaintextSecrets := kutil.HasMetaDataAnnotation(secret, common.EtcdEncryptionForcePlaintextAnnotationName, "true")

		encrypt := !firstCreationOfEncryptionConfiguration && !forcePlaintextSecrets
		b.Logger.Infof("Setting encryption of %s to %t", common.EtcdEncryptionEncryptedResourceSecrets, encrypt)
		if err := encryptionconfiguration.SetResourceEncryption(conf, common.EtcdEncryptionEncryptedResourceSecrets, encrypt); err != nil {
			return err
		}

		checksum, err := confChecksum(conf)
		if err != nil {
			return err
		}

		func() {
			b.mutex.Lock()
			defer b.mutex.Unlock()
			b.CheckSums[common.EtcdEncryptionSecretName] = checksum
		}()

		return encryptionconfiguration.UpdateSecret(secret, conf)
	})
	if err != nil {
		return nil, err
	}

	return conf, err
}

func (b *Botanist) syncEncryptionConfigurationToGarden(ctx context.Context, conf *apiserverconfigv1.EncryptionConfiguration) error {
	secret := &corev1.Secret{ObjectMeta: kutil.ObjectMetaFromKey(common.GardenEtcdEncryptionSecretKey(b.Shoot.Info.Namespace, b.Shoot.Info.Name))}
	_, err := controllerutil.CreateOrUpdate(ctx, b.K8sGardenClient.Client(), secret, func() error {
		return encryptionconfiguration.UpdateSecret(secret, conf)
	})
	return err
}

func confChecksum(conf *apiserverconfigv1.EncryptionConfiguration) (string, error) {
	data, err := encryptionconfiguration.Write(conf)
	if err != nil {
		return "", err
	}

	return utils.ComputeSHA256Hex(data), nil
}

// RewriteShootSecretsIfEncryptionConfigurationChanged rewrites the secrets in the Shoot if the etcd
// encryption configuration changed. Rewriting here means that a patch request is sent that forces
// the etcd to encrypt them with the new configuration.
func (b *Botanist) RewriteShootSecretsIfEncryptionConfigurationChanged(ctx context.Context) error {
	checksum := func() string {
		b.mutex.RLock()
		defer b.mutex.RUnlock()
		return b.CheckSums[common.EtcdEncryptionSecretName]
	}()

	secret := &corev1.Secret{}
	if err := b.K8sSeedClient.Client().Get(ctx, kutil.Key(b.Shoot.SeedNamespace, common.EtcdEncryptionSecretName), secret); err != nil {
		return err
	}

	if secret.Annotations[common.EtcdEncryptionChecksumAnnotationName] == checksum {
		b.Logger.Infof("etcd encryption is up to date (checksum %s), no need to rewrite secrets", checksum)
		return nil
	}

	shortChecksum := kutil.TruncateLabelValue(checksum)
	notCurrentChecksum, err := labels.NewRequirement(common.EtcdEncryptionChecksumLabelName, selection.NotEquals, []string{shortChecksum})
	if err != nil {
		return err
	}

	secretSelectOptions := client.UseListOptions(&client.ListOptions{
		LabelSelector: labels.NewSelector().Add(*notCurrentChecksum),
	})

	secretList := &corev1.SecretList{}
	if err = b.K8sShootClient.Client().List(ctx, secretList, secretSelectOptions); err != nil {
		return err
	}

	for _, s := range secretList.Items {
		withoutChecksumLabel := s.DeepCopy()
		kutil.SetMetaDataLabel(&s, common.EtcdEncryptionChecksumLabelName, shortChecksum)
		if err := b.K8sShootClient.Client().Patch(ctx, &s, client.MergeFrom(withoutChecksumLabel)); client.IgnoreNotFound(err) != nil {
			return err
		}
		b.Logger.Debugf("Successfully rewrote secret %v/%v (checksum %q)", s.Namespace, s.Name, shortChecksum)
	}

	oldSecret := secret.DeepCopy()
	kutil.SetMetaDataAnnotation(secret, common.EtcdEncryptionChecksumAnnotationName, checksum)
	return b.K8sSeedClient.Client().Patch(ctx, secret, client.MergeFrom(oldSecret))
}
