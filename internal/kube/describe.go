package kube

// EventInfo is used in PodDetail to surface kubernetes events for a pod.
type EventInfo struct {
	Type    string `json:"type"`
	Reason  string `json:"reason"`
	Age     string `json:"age"`
	From    string `json:"from"`
	Message string `json:"message"`
	Count   int32  `json:"count"`
}
