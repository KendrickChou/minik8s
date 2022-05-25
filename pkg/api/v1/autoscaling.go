package v1

import "time"

type HorizontalPodAutoscaler struct {
	TypeMeta
	ObjectMeta `json:"metadata"`
	Spec       HorizontalPodAutoscalerSpec   `json:"spec,omitempty"`
	Status     HorizontalPodAutoscalerStatus `json:"status,omitempty"`
}

type HorizontalPodAutoscalerSpec struct {
	ScaleTargetRef CrossVersionObjectReference      `json:"scaleTargetRef"`
	MinReplicas    int                              `json:"minReplicas,omitempty"`
	MaxReplicas    int                              `json:"maxReplicas,omitempty"`
	Metrics        []MetricSpec                     `json:"metrics,omitempty"`
	Behavior       *HorizontalPodAutoscalerBehavior `json:"behavior,omitempty"`
}

type CrossVersionObjectReference struct {
	Kind       string `json:"kind"`
	Name       string `json:"name"`
	APIVersion string `json:"apiVersion,omitempty"`
}

type MetricSpec struct {
	Type     MetricSourceType     `json:"type"`
	Resource ResourceMetricSource `json:"resource,omitempty"`
}

// MetricSourceType miniK8s only support "Resource"
type MetricSourceType string

const (
	ResourceMetricSourceType MetricSourceType = "Resource"
)

type ResourceMetricSource struct {
	Name   string       `json:"name"`
	Target MetricTarget `json:"target"`
}

// MetricTarget type can only be "Utilization" in miniK8s
type MetricTarget struct {
	Type MetricTargetType `json:"type"`
	//Value              *string          `json:"value,omitempty"`
	//AverageValue       *string          `json:"averageValue,omitempty"`
	AverageUtilization int `json:"averageUtilization,omitempty"`
}

// MetricTargetType specifies the type of metric being targeted, and should be either
// "Value", "AverageValue", or "Utilization"
type MetricTargetType string

const (
	UtilizationMetricType  MetricTargetType = "Utilization"
	ValueMetricType        MetricTargetType = "Value"
	AverageValueMetricType MetricTargetType = "AverageValue"
)

type HorizontalPodAutoscalerBehavior struct {
	ScaleUp   *HPAScalingRules `json:"scaleUp,omitempty"`
	ScaleDown *HPAScalingRules `json:"scaleDown,omitempty"`
}

type HPAScalingRules struct {
	// StabilizationWindowSeconds is the number of seconds for which past recommendations should be
	// considered while scaling up or scaling down.
	// StabilizationWindowSeconds must be greater than or equal to zero and less than or equal to 3600 (one hour).
	// If not set, use the default values:
	// - For scale up: 0 (i.e. no stabilization is done).
	// - For scale down: 300 (i.e. the stabilization window is 300 seconds long).
	// +optional
	StabilizationWindowSeconds int `json:"stabilizationWindowSeconds,omitempty"`
	// selectPolicy is used to specify which policy should be used.
	// If not set, the default value Max is used.
	// +optional
	SelectPolicy ScalingPolicySelect `json:"selectPolicy,omitempty"`
	// policies is a list of potential scaling polices which can be used during scaling.
	// At least one policy must be specified, otherwise the HPAScalingRules will be discarded as invalid
	// +listType=atomic
	// +optional
	Policies []HPAScalingPolicy `json:"policies,omitempty"`
}

// ScalingPolicySelect is used to specify which policy should be used while scaling in a certain direction
type ScalingPolicySelect string

const (
	// MaxChangePolicySelect  selects the policy with the highest possible change.
	MaxChangePolicySelect ScalingPolicySelect = "Max"
	// MinChangePolicySelect selects the policy with the lowest possible change.
	MinChangePolicySelect ScalingPolicySelect = "Min"
	// DisabledPolicySelect disables the scaling in this direction.
	DisabledPolicySelect ScalingPolicySelect = "Disabled"
)

// HPAScalingPolicyType is the type of the policy which could be used while making scaling decisions.
type HPAScalingPolicyType string

const (
	// PodsScalingPolicy is a policy used to specify a change in absolute number of pods.
	PodsScalingPolicy HPAScalingPolicyType = "Pods"
	// PercentScalingPolicy is a policy used to specify a relative amount of change with respect to
	// the current number of pods.
	PercentScalingPolicy HPAScalingPolicyType = "Percent"
)

// HPAScalingPolicy is a single policy which must hold true for a specified past interval.
type HPAScalingPolicy struct {
	// Type is used to specify the scaling policy.
	Type HPAScalingPolicyType `json:"type"`
	// Value contains the amount of change which is permitted by the policy.
	// It must be greater than zero
	Value int `json:"value"`
	// PeriodSeconds specifies the window of time for which the policy should hold true.
	// PeriodSeconds must be greater than zero and less than or equal to 1800 (30 min).
	PeriodSeconds int `json:"periodSeconds"`
}

type HorizontalPodAutoscalerStatus struct {
	LastScaleTime   time.Time `json:"lastScaleTime,omitempty"`
	CurrentReplicas int       `json:"currentReplicas,omitempty"`
	DesiredReplicas int       `json:"desiredReplicas,omitempty"`
}
