package k8s

import (
	"yager/model"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//"github.com/spf13/viper"
)

func scheduleAffinity(dts *model.App) *apiv1.Affinity {
	affinity := &apiv1.Affinity{
		PodAntiAffinity: &apiv1.PodAntiAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []apiv1.WeightedPodAffinityTerm{
				{
					// weight associated with matching the corresponding podAffinityTerm,
					// in the range 1-100.
					Weight: 100,
					PodAffinityTerm: apiv1.PodAffinityTerm{
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"cloud.graviti.cn/appid": dts.AppID,
							},
						},
						TopologyKey: "kubernetes.io/hostname",
					},
				},
			},
		},
	}

	var nsTerms []apiv1.NodeSelectorTerm

	nsTerms = append(nsTerms, apiv1.NodeSelectorTerm{
		MatchExpressions: []apiv1.NodeSelectorRequirement{
			{
				Key:      "cloud.graviti.cn/cni",
				Operator: apiv1.NodeSelectorOpIn,
				Values:   []string{"flannel"},
			},
		},
	})

	//nsTerms = append(nsTerms, cfgNodeSelectorTerms(grp)...)

	if len(nsTerms) != 0 {
		affinity.NodeAffinity = &apiv1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &apiv1.NodeSelector{
				NodeSelectorTerms: nsTerms,
			},
		}
	}

	return affinity
}

func scheduleToleration(dst *model.App) []apiv1.Toleration {

	var tols []apiv1.Toleration
	tols = append(tols, apiv1.Toleration{
		Key:      "node-role.kubernetes.io/master",
		Operator: apiv1.TolerationOpEqual,
		Value:    "",
		Effect:   apiv1.TaintEffectNoSchedule,
	})

	return tols
}
