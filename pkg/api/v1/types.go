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
