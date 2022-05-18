package v1

type LabelSelector struct {
	MatchLabels map[string]string `json:"matchLabels,omitempty"`
}

type PodTemplateSpec struct {
	ObjectMeta `json:"metadata,omitempty"`
	Spec       PodSpec `json:"spec,omitempty"`
}

type ConditionStatus string

const (
	ConditionTrue    ConditionStatus = "True"
	ConditionFalse   ConditionStatus = "False"
	ConditionUnknown ConditionStatus = "Unknown"
)

func MatchSelector(requirement LabelSelector, labels map[string]string) bool {
	flag := true
	for reqKey, reqVal := range requirement.MatchLabels {
		givenVal, exist := labels[reqKey]
		if !exist || reqVal != givenVal {
			flag = false
			break
		}
	}
	return flag
}

func MatchLabels(requirement map[string]string, labels map[string]string) bool {
	flag := true
	for reqKey, reqVal := range requirement {
		givenVal, exist := labels[reqKey]
		if !exist || reqVal != givenVal {
			flag = false
			break
		}
	}
	return flag
}

func CheckOwner(owners []OwnerReference, key string) int {
	for index, owner := range owners {
		if owner.UID == key {
			return index
		}
	}

	return -1
}

func ComparePodStatus(newStatus *PodStatus, oldStatus *PodStatus) bool {
	if newStatus.PodIP != oldStatus.PodIP {
		return false
	}

	if newStatus.PodNetworkID != oldStatus.PodNetworkID {
		return false
	}

	if newStatus.Phase != oldStatus.Phase {
		return false
	}

	return true
}

func GetOwnerReplicaSet(pod *Pod) string {
	for _, owner := range pod.OwnerReferences {
		if owner.Kind == "Replica" {
			return owner.UID
		}
	}

	return ""
}
