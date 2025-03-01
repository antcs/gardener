// Copyright (c) 2018 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
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

package helper_test

import (
	gardenv1beta1 "github.com/gardener/gardener/pkg/apis/garden/v1beta1"
	. "github.com/gardener/gardener/pkg/apis/garden/v1beta1/helper"
	"github.com/gardener/gardener/pkg/operation/common"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var (
	trueVar  = true
	falseVar = false
)

var _ = Describe("helper", func() {
	DescribeTable("#IsShootHibernated",
		func(shoot *gardenv1beta1.Shoot, hibernated bool) {
			Expect(IsShootHibernated(shoot)).To(Equal(hibernated))
		},
		Entry("no hibernation section", &gardenv1beta1.Shoot{}, false),
		Entry("hibernation.enabled = false", &gardenv1beta1.Shoot{
			Spec: gardenv1beta1.ShootSpec{
				Hibernation: &gardenv1beta1.Hibernation{Enabled: false},
			},
		}, false),
		Entry("hibernation.enabled = true", &gardenv1beta1.Shoot{
			Spec: gardenv1beta1.ShootSpec{
				Hibernation: &gardenv1beta1.Hibernation{Enabled: true},
			},
		}, true),
	)

	DescribeTable("#GetShootCloudProviderWorkers",
		func(cloudProvider gardenv1beta1.CloudProvider, shoot *gardenv1beta1.Shoot, expected []gardenv1beta1.Worker) {
			Expect(GetShootCloudProviderWorkers(cloudProvider, shoot)).To(Equal(expected))
		},
		Entry("AWS",
			gardenv1beta1.CloudProviderAWS,
			&gardenv1beta1.Shoot{
				Spec: gardenv1beta1.ShootSpec{
					Cloud: gardenv1beta1.Cloud{
						AWS: &gardenv1beta1.AWSCloud{
							Workers: []gardenv1beta1.AWSWorker{{Worker: gardenv1beta1.Worker{Name: "aws"}}},
						},
					},
				},
			},
			[]gardenv1beta1.Worker{{Name: "aws"}}),
		Entry("Azure",
			gardenv1beta1.CloudProviderAzure,
			&gardenv1beta1.Shoot{
				Spec: gardenv1beta1.ShootSpec{
					Cloud: gardenv1beta1.Cloud{
						Azure: &gardenv1beta1.AzureCloud{
							Workers: []gardenv1beta1.AzureWorker{{Worker: gardenv1beta1.Worker{Name: "azure"}}},
						},
					},
				},
			},
			[]gardenv1beta1.Worker{{Name: "azure"}}),
		Entry("GCP",
			gardenv1beta1.CloudProviderGCP,
			&gardenv1beta1.Shoot{
				Spec: gardenv1beta1.ShootSpec{
					Cloud: gardenv1beta1.Cloud{
						GCP: &gardenv1beta1.GCPCloud{
							Workers: []gardenv1beta1.GCPWorker{{Worker: gardenv1beta1.Worker{Name: "gcp"}}},
						},
					},
				},
			},
			[]gardenv1beta1.Worker{{Name: "gcp"}}),
		Entry("OpenStack",
			gardenv1beta1.CloudProviderOpenStack,
			&gardenv1beta1.Shoot{
				Spec: gardenv1beta1.ShootSpec{
					Cloud: gardenv1beta1.Cloud{
						OpenStack: &gardenv1beta1.OpenStackCloud{
							Workers: []gardenv1beta1.OpenStackWorker{{Worker: gardenv1beta1.Worker{Name: "openStack"}}},
						},
					},
				},
			},
			[]gardenv1beta1.Worker{{Name: "openStack"}}),
	)

	var (
		machineImageName    = "some-machineImage"
		machineImageVersion = "some-version"
		machineImage        = &gardenv1beta1.MachineImage{
			Name:    machineImageName,
			Version: machineImageVersion,
		}
	)

	DescribeTable("#GetMachineImageFromShoot",
		func(cloudProvider gardenv1beta1.CloudProvider, shoot *gardenv1beta1.Shoot, expected *gardenv1beta1.MachineImage) {
			Expect(GetMachineImageFromShoot(cloudProvider, shoot)).To(Equal(expected))
		},
		Entry("AWS",
			gardenv1beta1.CloudProviderAWS,
			&gardenv1beta1.Shoot{
				Spec: gardenv1beta1.ShootSpec{
					Cloud: gardenv1beta1.Cloud{
						AWS: &gardenv1beta1.AWSCloud{
							MachineImage: machineImage,
						},
					},
				},
			},
			machineImage,
		),
		Entry("Azure",
			gardenv1beta1.CloudProviderAzure,
			&gardenv1beta1.Shoot{
				Spec: gardenv1beta1.ShootSpec{
					Cloud: gardenv1beta1.Cloud{
						Azure: &gardenv1beta1.AzureCloud{
							MachineImage: machineImage,
						},
					},
				},
			},
			machineImage,
		),
		Entry("GCP",
			gardenv1beta1.CloudProviderGCP,
			&gardenv1beta1.Shoot{
				Spec: gardenv1beta1.ShootSpec{
					Cloud: gardenv1beta1.Cloud{
						GCP: &gardenv1beta1.GCPCloud{
							MachineImage: machineImage,
						},
					},
				},
			},
			machineImage,
		),
		Entry("OpenStack",
			gardenv1beta1.CloudProviderOpenStack,
			&gardenv1beta1.Shoot{
				Spec: gardenv1beta1.ShootSpec{
					Cloud: gardenv1beta1.Cloud{
						OpenStack: &gardenv1beta1.OpenStackCloud{
							MachineImage: machineImage,
						},
					},
				},
			},
			machineImage,
		),
		Entry("Alicloud",
			gardenv1beta1.CloudProviderAlicloud,
			&gardenv1beta1.Shoot{
				Spec: gardenv1beta1.ShootSpec{
					Cloud: gardenv1beta1.Cloud{
						Alicloud: &gardenv1beta1.Alicloud{
							MachineImage: machineImage,
						},
					},
				},
			},
			machineImage,
		),
		Entry("Packet",
			gardenv1beta1.CloudProviderPacket,
			&gardenv1beta1.Shoot{
				Spec: gardenv1beta1.ShootSpec{
					Cloud: gardenv1beta1.Cloud{
						Packet: &gardenv1beta1.PacketCloud{
							MachineImage: machineImage,
						},
					},
				},
			},
			machineImage,
		),
	)

	var (
		kubernetesConstraint = gardenv1beta1.KubernetesConstraints{
			Versions: []string{
				"1.15.1",
				"1.14.4",
				"1.12.9",
			},
		}
	)

	DescribeTable("#DetermineLatestKubernetesPatchVersion",
		func(cloudProfile gardenv1beta1.CloudProfile, currentVersion, expectedVersion string, expectVersion bool) {
			ok, newVersion, err := DetermineLatestKubernetesPatchVersion(cloudProfile, currentVersion)
			Expect(err).ToNot(HaveOccurred())
			Expect(ok).To(Equal(expectVersion))
			Expect(newVersion).To(Equal(expectedVersion))
		},
		Entry("version = 1.15.1",
			gardenv1beta1.CloudProfile{
				Spec: gardenv1beta1.CloudProfileSpec{
					AWS: &gardenv1beta1.AWSProfile{
						Constraints: gardenv1beta1.AWSConstraints{
							Kubernetes: kubernetesConstraint,
						},
					},
				},
			},
			"1.15.0",
			"1.15.1",
			true,
		),
		Entry("version = 1.12.9",
			gardenv1beta1.CloudProfile{
				Spec: gardenv1beta1.CloudProfileSpec{
					AWS: &gardenv1beta1.AWSProfile{
						Constraints: gardenv1beta1.AWSConstraints{
							Kubernetes: kubernetesConstraint,
						},
					},
				},
			},
			"1.12.4",
			"1.12.9",
			true,
		),
		Entry("no new version",
			gardenv1beta1.CloudProfile{
				Spec: gardenv1beta1.CloudProfileSpec{
					AWS: &gardenv1beta1.AWSProfile{
						Constraints: gardenv1beta1.AWSConstraints{
							Kubernetes: kubernetesConstraint,
						},
					},
				},
			},
			"1.15.1",
			"",
			false,
		),
		Entry("GCP",
			gardenv1beta1.CloudProfile{
				Spec: gardenv1beta1.CloudProfileSpec{
					GCP: &gardenv1beta1.GCPProfile{
						Constraints: gardenv1beta1.GCPConstraints{
							Kubernetes: kubernetesConstraint,
						},
					},
				},
			},
			"1.12.4",
			"1.12.9",
			true,
		),
		Entry("Azure",
			gardenv1beta1.CloudProfile{
				Spec: gardenv1beta1.CloudProfileSpec{
					Azure: &gardenv1beta1.AzureProfile{
						Constraints: gardenv1beta1.AzureConstraints{
							Kubernetes: kubernetesConstraint,
						},
					},
				},
			},
			"1.12.4",
			"1.12.9",
			true,
		),
		Entry("Openstack",
			gardenv1beta1.CloudProfile{
				Spec: gardenv1beta1.CloudProfileSpec{
					OpenStack: &gardenv1beta1.OpenStackProfile{
						Constraints: gardenv1beta1.OpenStackConstraints{
							Kubernetes: kubernetesConstraint,
						},
					},
				},
			},
			"1.12.4",
			"1.12.9",
			true,
		),
		Entry("Packet",
			gardenv1beta1.CloudProfile{
				Spec: gardenv1beta1.CloudProfileSpec{
					Packet: &gardenv1beta1.PacketProfile{
						Constraints: gardenv1beta1.PacketConstraints{
							Kubernetes: kubernetesConstraint,
						},
					},
				},
			},
			"1.12.4",
			"1.12.9",
			true,
		),
	)

	DescribeTable("#DetermineNextKubernetesMinorVersion",
		func(cloudProfile gardenv1beta1.CloudProfile, currentVersion, expectedVersion string, expectVersion bool) {
			ok, newVersion, err := DetermineNextKubernetesMinorVersion(cloudProfile, currentVersion)
			Expect(err).ToNot(HaveOccurred())
			Expect(ok).To(Equal(expectVersion))
			Expect(newVersion).To(Equal(expectedVersion))
		},
		Entry("version = 1.15.1",
			gardenv1beta1.CloudProfile{
				Spec: gardenv1beta1.CloudProfileSpec{
					AWS: &gardenv1beta1.AWSProfile{
						Constraints: gardenv1beta1.AWSConstraints{
							Kubernetes: kubernetesConstraint,
						},
					},
				},
			},
			"1.14.4",
			"1.15.1",
			true,
		),
		Entry("version = 1.12.9",
			gardenv1beta1.CloudProfile{
				Spec: gardenv1beta1.CloudProfileSpec{
					AWS: &gardenv1beta1.AWSProfile{
						Constraints: gardenv1beta1.AWSConstraints{
							Kubernetes: kubernetesConstraint,
						},
					},
				},
			},
			"1.11.0",
			"1.12.9",
			true,
		),
		Entry("no new version",
			gardenv1beta1.CloudProfile{
				Spec: gardenv1beta1.CloudProfileSpec{
					AWS: &gardenv1beta1.AWSProfile{
						Constraints: gardenv1beta1.AWSConstraints{
							Kubernetes: kubernetesConstraint,
						},
					},
				},
			},
			"1.15.1",
			"",
			false,
		),
		Entry("GCP",
			gardenv1beta1.CloudProfile{
				Spec: gardenv1beta1.CloudProfileSpec{
					GCP: &gardenv1beta1.GCPProfile{
						Constraints: gardenv1beta1.GCPConstraints{
							Kubernetes: kubernetesConstraint,
						},
					},
				},
			},
			"1.12.9",
			"1.14.4",
			true,
		),
		Entry("Azure",
			gardenv1beta1.CloudProfile{
				Spec: gardenv1beta1.CloudProfileSpec{
					Azure: &gardenv1beta1.AzureProfile{
						Constraints: gardenv1beta1.AzureConstraints{
							Kubernetes: kubernetesConstraint,
						},
					},
				},
			},
			"1.12.9",
			"1.14.4",
			true,
		),
		Entry("Openstack",
			gardenv1beta1.CloudProfile{
				Spec: gardenv1beta1.CloudProfileSpec{
					OpenStack: &gardenv1beta1.OpenStackProfile{
						Constraints: gardenv1beta1.OpenStackConstraints{
							Kubernetes: kubernetesConstraint,
						},
					},
				},
			},
			"1.12.9",
			"1.14.4",
			true,
		),
		Entry("Packet",
			gardenv1beta1.CloudProfile{
				Spec: gardenv1beta1.CloudProfileSpec{
					Packet: &gardenv1beta1.PacketProfile{
						Constraints: gardenv1beta1.PacketConstraints{
							Kubernetes: kubernetesConstraint,
						},
					},
				},
			},
			"1.12.9",
			"1.14.4",
			true,
		),
	)

	DescribeTable("#ShootWantsClusterAutoscaler",
		func(shoot *gardenv1beta1.Shoot, wantsAutoscaler bool) {
			actualWantsAutoscaler, err := ShootWantsClusterAutoscaler(shoot)

			Expect(err).NotTo(HaveOccurred())
			Expect(actualWantsAutoscaler).To(Equal(wantsAutoscaler))
		},
		Entry("no workers",
			&gardenv1beta1.Shoot{
				Spec: gardenv1beta1.ShootSpec{
					Cloud: gardenv1beta1.Cloud{
						OpenStack: &gardenv1beta1.OpenStackCloud{},
					},
				},
			},
			false),
		Entry("one worker no difference in auto scaler max and min",
			&gardenv1beta1.Shoot{
				Spec: gardenv1beta1.ShootSpec{
					Cloud: gardenv1beta1.Cloud{
						OpenStack: &gardenv1beta1.OpenStackCloud{
							Workers: []gardenv1beta1.OpenStackWorker{{Worker: gardenv1beta1.Worker{Name: "foo"}}},
						},
					},
				},
			},
			false),
		Entry("one worker with difference in auto scaler max and min",
			&gardenv1beta1.Shoot{
				Spec: gardenv1beta1.ShootSpec{
					Cloud: gardenv1beta1.Cloud{
						OpenStack: &gardenv1beta1.OpenStackCloud{
							Workers: []gardenv1beta1.OpenStackWorker{{Worker: gardenv1beta1.Worker{Name: "foo", AutoScalerMin: 1, AutoScalerMax: 2}}},
						},
					},
				},
			},
			true))

	var (
		alertingSecrets = map[string]*corev1.Secret{
			common.GardenRoleAlertingSMTP: {},
		}
	)

	DescribeTable("#ShootWantsAlertmanager",
		func(shoot *gardenv1beta1.Shoot, secrets map[string]*corev1.Secret, wantsAlertmanager bool) {
			actualWantsAlertmanager := ShootWantsAlertmanager(shoot, secrets)
			Expect(actualWantsAlertmanager).To(Equal(wantsAlertmanager))
		},
		Entry("alertmanger wanted", &gardenv1beta1.Shoot{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					common.GardenOperatedBy: "test@gardener.cloud",
				},
			},
		}, alertingSecrets, true),
		Entry("no alertmanager due to missing smtp secret", &gardenv1beta1.Shoot{}, map[string]*corev1.Secret{}, false),
		Entry("no alertmanager due to missing operatedBy annotation", &gardenv1beta1.Shoot{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{},
			},
		}, alertingSecrets, false),
		Entry("no alertmanager wanted due to invalid mail address", &gardenv1beta1.Shoot{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					common.GardenOperatedBy: "invalid-mail-address",
				},
			},
		}, alertingSecrets, false))

	Describe("#ReadShootedSeed", func() {
		var (
			shoot                    *gardenv1beta1.Shoot
			defaultReplicas          int32 = 3
			defaultMinReplicas       int32 = 3
			defaultMaxReplicas       int32 = 3
			defaultMinimumVolumeSize       = "20Gi"

			defaultAPIServerAutoscaler = ShootedSeedAPIServerAutoscaler{
				MinReplicas: &defaultMinReplicas,
				MaxReplicas: defaultMaxReplicas,
			}

			defaultAPIServer = ShootedSeedAPIServer{
				Replicas:   &defaultReplicas,
				Autoscaler: &defaultAPIServerAutoscaler,
			}

			defaultShootedSeed = ShootedSeed{
				APIServer: &defaultAPIServer,
			}
		)

		BeforeEach(func() {
			shoot = &gardenv1beta1.Shoot{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:   common.GardenNamespace,
					Annotations: nil,
				},
			}
		})

		It("should return false,nil,nil because shoot is not in the garden namespace", func() {
			shoot.Namespace = "default"

			shootedSeed, err := ReadShootedSeed(shoot)

			Expect(err).NotTo(HaveOccurred())
			Expect(shootedSeed).To(BeNil())
		})

		It("should return false,nil,nil because annotation is not set", func() {
			shootedSeed, err := ReadShootedSeed(shoot)

			Expect(err).NotTo(HaveOccurred())
			Expect(shootedSeed).To(BeNil())
		})

		It("should return false,nil,nil because annotation is set with no usages", func() {
			shoot.Annotations = map[string]string{
				common.ShootUseAsSeed: "",
			}

			shootedSeed, err := ReadShootedSeed(shoot)

			Expect(err).NotTo(HaveOccurred())
			Expect(shootedSeed).To(BeNil())
		})

		It("should return true,nil,nil because annotation is set with normal usage", func() {
			shoot.Annotations = map[string]string{
				common.ShootUseAsSeed: "true",
			}

			shootedSeed, err := ReadShootedSeed(shoot)

			Expect(err).NotTo(HaveOccurred())
			Expect(shootedSeed).To(Equal(&defaultShootedSeed))
		})

		It("should return true,true,true because annotation is set with protected and visible usage", func() {
			shoot.Annotations = map[string]string{
				common.ShootUseAsSeed: "true,protected,visible",
			}

			shootedSeed, err := ReadShootedSeed(shoot)

			Expect(err).NotTo(HaveOccurred())
			Expect(shootedSeed).To(Equal(&ShootedSeed{
				Protected: &trueVar,
				Visible:   &trueVar,
				APIServer: &defaultAPIServer,
			}))
		})

		It("should return true,true,true because annotation is set with unprotected and invisible usage", func() {
			shoot.Annotations = map[string]string{
				common.ShootUseAsSeed: "true,unprotected,invisible",
			}

			shootedSeed, err := ReadShootedSeed(shoot)

			Expect(err).NotTo(HaveOccurred())
			Expect(shootedSeed).To(Equal(&ShootedSeed{
				Protected:         &falseVar,
				Visible:           &falseVar,
				APIServer:         &defaultAPIServer,
				MinimumVolumeSize: nil,
			}))
		})

		It("should return the min volume size because annotation is set properly", func() {
			shoot.Annotations = map[string]string{
				common.ShootUseAsSeed: "true,unprotected,invisible,minimumVolumeSize=20Gi",
			}

			shootedSeed, err := ReadShootedSeed(shoot)

			Expect(err).NotTo(HaveOccurred())
			Expect(shootedSeed).To(Equal(&ShootedSeed{
				Protected:         &falseVar,
				Visible:           &falseVar,
				APIServer:         &defaultAPIServer,
				MinimumVolumeSize: &defaultMinimumVolumeSize,
			}))
		})

		It("should return a filled apiserver config", func() {
			shoot.Annotations = map[string]string{
				common.ShootUseAsSeed: "true,apiServer.replicas=1,apiServer.autoscaler.minReplicas=2,apiServer.autoscaler.maxReplicas=3",
			}

			shootedSeed, err := ReadShootedSeed(shoot)

			var (
				one   int32 = 1
				two   int32 = 2
				three int32 = 3
			)

			Expect(err).NotTo(HaveOccurred())
			Expect(shootedSeed).To(Equal(&ShootedSeed{
				APIServer: &ShootedSeedAPIServer{
					Replicas: &one,
					Autoscaler: &ShootedSeedAPIServerAutoscaler{
						MinReplicas: &two,
						MaxReplicas: three,
					},
				},
			}))
		})

		It("should fail due to maxReplicas not being specified", func() {
			shoot.Annotations = map[string]string{
				common.ShootUseAsSeed: "true,apiServer.autoscaler.minReplicas=2",
			}

			_, err := ReadShootedSeed(shoot)
			Expect(err).To(HaveOccurred())
		})

		It("should fail due to API server replicas being less than one", func() {
			shoot.Annotations = map[string]string{
				common.ShootUseAsSeed: "true,apiServer.replicas=0",
			}

			_, err := ReadShootedSeed(shoot)
			Expect(err).To(HaveOccurred())
		})

		It("should fail due to API server autoscaler minReplicas being less than one", func() {
			shoot.Annotations = map[string]string{
				common.ShootUseAsSeed: "true,apiServer.autoscaler.minReplicas=0,apiServer.autoscaler.maxReplicas=1",
			}

			_, err := ReadShootedSeed(shoot)
			Expect(err).To(HaveOccurred())
		})

		It("should fail due to API server autoscaler maxReplicas being less than one", func() {
			shoot.Annotations = map[string]string{
				common.ShootUseAsSeed: "true,apiServer.autoscaler.maxReplicas=0",
			}

			_, err := ReadShootedSeed(shoot)
			Expect(err).To(HaveOccurred())
		})

		It("should fail due to API server autoscaler minReplicas being greater than maxReplicas", func() {
			shoot.Annotations = map[string]string{
				common.ShootUseAsSeed: "true,apiServer.autoscaler.maxReplicas=1,apiServer.autoscaler.minReplicas=2",
			}

			_, err := ReadShootedSeed(shoot)
			Expect(err).To(HaveOccurred())
		})
	})
})
