// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package omni provides resources describing the Machines, Clusters, etc.
package omni

import "github.com/siderolabs/omni/client/pkg/omni/resources/registry"

func init() {
	registry.MustRegisterResource(ClusterBootstrapStatusType, &ClusterBootstrapStatus{})
	registry.MustRegisterResource(ClusterConfigVersionType, &ClusterConfigVersion{})
	registry.MustRegisterResource(ClusterEndpointType, &ClusterEndpoint{})
	registry.MustRegisterResource(ClusterDestroyStatusType, &ClusterDestroyStatus{})
	registry.MustRegisterResource(ClusterWorkloadProxyStatusType, &ClusterWorkloadProxyStatus{})
	registry.MustRegisterResource(ClusterKubernetesNodesType, &ClusterKubernetesNodes{})
	registry.MustRegisterResource(ClusterType, &Cluster{})
	registry.MustRegisterResource(ClusterUUIDType, &ClusterUUID{})
	registry.MustRegisterResource(ClusterSecretsType, &ClusterSecrets{})
	registry.MustRegisterResource(ClusterStatusType, &ClusterStatus{})
	registry.MustRegisterResource(ClusterMachineType, &ClusterMachine{})
	registry.MustRegisterResource(ClusterMachineEncryptionKeyType, &ClusterMachineEncryptionKey{})
	registry.MustRegisterResource(ClusterMachineIdentityType, &ClusterMachineIdentity{})
	registry.MustRegisterResource(ClusterMachineStatusType, &ClusterMachineStatus{})
	registry.MustRegisterResource(ClusterMachineConfigType, &ClusterMachineConfig{})
	registry.MustRegisterResource(ClusterMachineConfigPatchesType, &ClusterMachineConfigPatches{})
	registry.MustRegisterResource(ClusterMachineConfigStatusType, &ClusterMachineConfigStatus{})
	registry.MustRegisterResource(ClusterMachineTalosVersionType, &ClusterMachineTalosVersion{})
	registry.MustRegisterResource(ClusterMachineTemplateType, &ClusterMachineTemplate{})
	registry.MustRegisterResource(ClusterTaintType, &ClusterTaint{})
	registry.MustRegisterResource(ConfigPatchType, &ConfigPatch{})
	registry.MustRegisterResource(EtcdAuditResultType, &EtcdAuditResult{})
	registry.MustRegisterResource(EtcdBackupType, &EtcdBackup{})
	registry.MustRegisterResource(EtcdBackupS3ConfType, &EtcdBackupS3Conf{})
	registry.MustRegisterResource(EtcdBackupEncryptionType, &EtcdBackupEncryption{})
	registry.MustRegisterResource(EtcdBackupStatusType, &EtcdBackupStatus{})
	registry.MustRegisterResource(EtcdBackupOverallStatusType, &EtcdBackupOverallStatus{})
	registry.MustRegisterResource(EtcdBackupStoreStatusType, &EtcdBackupStoreStatus{})
	registry.MustRegisterResource(ExtensionsConfigurationType, &ExtensionsConfiguration{})
	registry.MustRegisterResource(ExtensionsConfigurationStatusType, &ExtensionsConfigurationStatus{})
	registry.MustRegisterResource(BackupDataType, &BackupData{})
	registry.MustRegisterResource(EtcdManualBackupType, &EtcdManualBackup{})
	registry.MustRegisterResource(ExposedServiceType, &ExposedService{})
	registry.MustRegisterResource(FeaturesConfigType, &FeaturesConfig{})
	registry.MustRegisterResource(ImagePullRequestType, &ImagePullRequest{})
	registry.MustRegisterResource(ImagePullStatusType, &ImagePullStatus{})
	registry.MustRegisterResource(InstallationMediaType, &InstallationMedia{})
	registry.MustRegisterResource(ControlPlaneStatusType, &ControlPlaneStatus{})
	registry.MustRegisterResource(KubeconfigType, &Kubeconfig{})
	registry.MustRegisterResource(KubernetesNodeAuditResultType, &KubernetesNodeAuditResult{})
	registry.MustRegisterResource(KubernetesStatusType, &KubernetesStatus{})
	registry.MustRegisterResource(KubernetesUpgradeManifestStatusType, &KubernetesUpgradeManifestStatus{})
	registry.MustRegisterResource(KubernetesUpgradeStatusType, &KubernetesUpgradeStatus{})
	registry.MustRegisterResource(KubernetesVersionType, &KubernetesVersion{})
	registry.MustRegisterResource(MachineLabelsType, &MachineLabels{})
	registry.MustRegisterResource(MachineType, &Machine{})
	registry.MustRegisterResource(MachineClassType, &MachineClass{})
	registry.MustRegisterResource(MachineClassStatusType, &MachineClassStatus{})
	registry.MustRegisterResource(MachineConfigGenOptionsType, &MachineConfigGenOptions{})
	registry.MustRegisterResource(MachineExtensionsStatusType, &MachineExtensionsStatus{})
	registry.MustRegisterResource(MachineExtensionsType, &MachineExtensions{})
	registry.MustRegisterResource(MachineProvisionType, &MachineProvision{})
	registry.MustRegisterResource(MachineRequestSetType, &MachineRequestSet{})
	registry.MustRegisterResource(MachineRequestSetStatusType, &MachineRequestSetStatus{})
	registry.MustRegisterResource(MachineRequestSetPressureType, &MachineRequestSetPressure{})
	registry.MustRegisterResource(MachineSetType, &MachineSet{})
	registry.MustRegisterResource(MachineSetDestroyStatusType, &MachineSetDestroyStatus{})
	registry.MustRegisterResource(MachineSetNodeType, &MachineSetNode{})
	registry.MustRegisterResource(MachineSetRequiredMachinesType, &MachineSetRequiredMachines{})
	registry.MustRegisterResource(MachineSetStatusType, &MachineSetStatus{})
	registry.MustRegisterResource(MachineStatusType, &MachineStatus{})
	registry.MustRegisterResource(MachineStatusSnapshotType, &MachineStatusSnapshot{})
	registry.MustRegisterResource(MachineStatusLinkType, &MachineStatusLink{})
	registry.MustRegisterResource(MachineStatusMetricsType, &MachineStatusMetrics{})
	registry.MustRegisterResource(LoadBalancerConfigType, &LoadBalancerConfig{})
	registry.MustRegisterResource(LoadBalancerStatusType, &LoadBalancerStatus{})
	registry.MustRegisterResource(OngoingTaskType, &OngoingTask{})
	registry.MustRegisterResource(RedactedClusterMachineConfigType, &RedactedClusterMachineConfig{})
	registry.MustRegisterResource(SchematicType, &Schematic{})
	registry.MustRegisterResource(SchematicConfigurationType, &SchematicConfiguration{})
	registry.MustRegisterResource(TalosConfigType, &TalosConfig{})
	registry.MustRegisterResource(TalosExtensionsType, &TalosExtensions{})
	registry.MustRegisterResource(TalosVersionType, &TalosVersion{})
	registry.MustRegisterResource(TalosUpgradeStatusType, &TalosUpgradeStatus{})
}
