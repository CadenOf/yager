package k8s

import (
	"voyager/model"
	VOL_CONST "voyager/pkg/constvar"

	apiv1 "k8s.io/api/core/v1"
)

func createVolumes(appMeta *model.AppMetaInfo) ([]apiv1.Volume, []apiv1.VolumeMount, error) {

	// init container volumes
	var volumeMount []apiv1.VolumeMount
	mountPropagationHostToContainer := apiv1.MountPropagationHostToContainer
	volumeMount = append(volumeMount, apiv1.VolumeMount{
		Name:      VOL_CONST.K8S_VOLUME_Name,
		MountPath: VOL_CONST.K8S_VOLUME_ContainerRootPath,
		// SubPath:          VOL_CONST.K8S_VOLUME_SubPath,
		SubPathExpr:      "$(" + VOL_CONST.K8S_ENV_PAAS_POD_NAME + ")",
		MountPropagation: &mountPropagationHostToContainer,
	})

	// init specVolumes
	var specVolumes []apiv1.Volume
	hostPathDirectoryOrCreate := apiv1.HostPathDirectoryOrCreate
	specVolumes = append(specVolumes, apiv1.Volume{
		Name: VOL_CONST.K8S_VOLUME_Name,
		VolumeSource: apiv1.VolumeSource{
			HostPath: &apiv1.HostPathVolumeSource{
				Path: VOL_CONST.K8S_VOLUME_RootPath + appMeta.Namespace + "/" + appMeta.AppID + "/" + appMeta.Name,
				Type: &hostPathDirectoryOrCreate,
			},
		},
	})

	return specVolumes, volumeMount, nil
}
