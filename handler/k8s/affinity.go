package k8s

import (
	"voyager/model"
	CONST "voyager/pkg/constvar"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Setting podAntiAffinity & nodeAffinity
func scheduleAffinity(aff *model.AffinityInfo) *apiv1.Affinity {
	// pod anti-affinity scheduling rules , avoid putting the same pods in the same node
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
								CONST.K8S_RESOURCE_ANNOTATION_appid: aff.AffMeta.AppID,
							},
						},
						TopologyKey: "kubernetes.io/hostname",
					},
				},
			},
		},
	}

	// nodeAffinity
	var nsTerms []apiv1.NodeSelectorTerm
	for _, ns := range aff.Selector {
		nsTerms = append(nsTerms, apiv1.NodeSelectorTerm{
			MatchExpressions: []apiv1.NodeSelectorRequirement{
				{
					Key:      ns.Key,
					Operator: apiv1.NodeSelectorOpIn,
					Values:   ns.Values,
				},
			},
		})
	}
	if len(nsTerms) != 0 {
		affinity.NodeAffinity = &apiv1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &apiv1.NodeSelector{
				NodeSelectorTerms: nsTerms,
			},
		}
	}

	return affinity
}

func scheduleToleration(tol *model.TolerationInfo) []apiv1.Toleration {

	var tols []apiv1.Toleration
	for _, tl := range tol.Toleration {
		tols = append(tols, apiv1.Toleration{
			Key:      tl.Key,
			Operator: apiv1.TolerationOpEqual,
			Value:    tl.Value,
			Effect:   apiv1.TaintEffectNoSchedule,
		})
	}
	return tols
}
