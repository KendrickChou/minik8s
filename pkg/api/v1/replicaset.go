package v1

import (
	"time"
)

type ReplicaSet struct {
	TypeMeta   `json:",inline"`
	ObjectMeta `json:"metadata,omitempty"`
	Spec       ReplicaSetSpec   `json:"spec,omitempty"`
	Status     ReplicaSetStatus `json:"status,omitempty"`
}

type ReplicaSetSpec struct {
	Replicas int32           `json:"replicas,omitempty"`
	Selector LabelSelector   `json:"selector"`
	Template PodTemplateSpec `json:"template,omitempty"`
}

// ReplicaSetStatus represents the current status of a ReplicaSet.
type ReplicaSetStatus struct {
	// Replicas is the most recently oberved number of replicas.
	Replicas int32 `json:"replicas"`

	// The number of pods that have labels matching the labels of the pod template of the replicaset.
	// +optional
	FullyLabeledReplicas int32 `json:"fullyLabeledReplicas,omitempty"`

	// readyReplicas is the number of pods targeted by this ReplicaSet with a Ready Condition.
	// +optional
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`

	// The number of available replicas (ready for at least minReadySeconds) for this replica set.
	// +optional
	AvailableReplicas int32 `json:"availableReplicas,omitempty"`

	// Represents the latest available observations of a replica set's current state.
	// +optional
	Conditions []ReplicaSetCondition `json:"conditions,omitempty"`
}

type ReplicaSetConditionType string

// ReplicaSetCondition describes the state of a replica set at a certain point.
type ReplicaSetCondition struct {
	// Type of replica set condition.
	Type ReplicaSetConditionType `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status ConditionStatus `json:"status"`
	// The last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime time.Time `json:"lastTransitionTime,omitempty"`
	// The reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty"`
	// A human readable message indicating details about the transition.
	// +optional
	Message string `json:"message,omitempty"`
}
