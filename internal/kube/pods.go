package kube

import (
	"context"
	"fmt"
	"sort"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PodSummary struct {
	Name      string `json:"name"`
	Status    string `json:"status"`
	Ready     string `json:"ready"`
	Restarts  int32  `json:"restarts"`
	Age       string `json:"age"`
	CreatedAt string `json:"createdAt"`
	Node      string `json:"node"`
	IP        string `json:"ip"`
}

type PodDetail struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Node        string            `json:"node"`
	IP          string            `json:"ip"`
	Status      string            `json:"status"`
	Phase       string            `json:"phase"`
	Age         string            `json:"age"`
	CreatedAt   string            `json:"createdAt"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	Containers  []ContainerDetail `json:"containers"`
	Conditions  []ConditionInfo   `json:"conditions"`
	Events      []EventInfo       `json:"events"`
	Owners      []RelatedResource `json:"owners,omitempty"`
}

type ContainerDetail struct {
	Name     string     `json:"name"`
	Image    string     `json:"image"`
	Ready    bool       `json:"ready"`
	Restarts int32      `json:"restarts"`
	State    string     `json:"state"`
	Ports    []PortInfo `json:"ports,omitempty"`
}

type PortInfo struct {
	Name          string `json:"name,omitempty"`
	ContainerPort int32  `json:"containerPort"`
	Protocol      string `json:"protocol"`
}

type ConditionInfo struct {
	Type    string `json:"type"`
	Status  string `json:"status"`
	Reason  string `json:"reason,omitempty"`
	Message string `json:"message,omitempty"`
}

func ListPods(ctx context.Context, contextName, namespace string) ([]PodSummary, error) {
	client, err := kubeClient(contextName)
	if err != nil {
		return nil, err
	}

	list, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	pods := make([]PodSummary, 0, len(list.Items))
	for _, p := range list.Items {
		status := string(p.Status.Phase)
		if p.DeletionTimestamp != nil {
			status = "Terminating"
		} else {
			for _, cs := range p.Status.ContainerStatuses {
				if cs.State.Waiting != nil && cs.State.Waiting.Reason != "" {
					status = cs.State.Waiting.Reason
					break
				}
			}
		}

		ready := 0
		for _, cs := range p.Status.ContainerStatuses {
			if cs.Ready {
				ready++
			}
		}

		var restarts int32
		for _, cs := range p.Status.ContainerStatuses {
			restarts += cs.RestartCount
		}

		created := ""
		if !p.CreationTimestamp.IsZero() {
			created = p.CreationTimestamp.UTC().Format(time.RFC3339)
		}

		pods = append(pods, PodSummary{
			Name:      p.Name,
			Status:    status,
			Ready:     fmt.Sprintf("%d/%d", ready, len(p.Spec.Containers)),
			Restarts:  restarts,
			Age:       podAge(p.CreationTimestamp.Time),
			CreatedAt: created,
			Node:      p.Spec.NodeName,
			IP:        p.Status.PodIP,
		})
	}

	sort.Slice(pods, func(i, j int) bool {
		return pods[i].CreatedAt < pods[j].CreatedAt
	})

	return pods, nil
}

func GetPod(ctx context.Context, contextName, namespace, name string) (*PodDetail, error) {
	client, err := kubeClient(contextName)
	if err != nil {
		return nil, err
	}

	p, err := client.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	containers := make([]ContainerDetail, 0, len(p.Spec.Containers))
	for _, c := range p.Spec.Containers {
		cd := ContainerDetail{Name: c.Name, Image: c.Image}
		for _, cs := range p.Status.ContainerStatuses {
			if cs.Name == c.Name {
				cd.Ready = cs.Ready
				cd.Restarts = cs.RestartCount
				switch {
				case cs.State.Running != nil:
					cd.State = "Running"
				case cs.State.Waiting != nil:
					cd.State = cs.State.Waiting.Reason
					if cd.State == "" {
						cd.State = "Waiting"
					}
				case cs.State.Terminated != nil:
					cd.State = cs.State.Terminated.Reason
					if cd.State == "" {
						cd.State = "Terminated"
					}
				default:
					cd.State = "Unknown"
				}
				break
			}
		}
		for _, port := range c.Ports {
			cd.Ports = append(cd.Ports, PortInfo{
				Name:          port.Name,
				ContainerPort: port.ContainerPort,
				Protocol:      string(port.Protocol),
			})
		}
		containers = append(containers, cd)
	}

	conditions := make([]ConditionInfo, 0, len(p.Status.Conditions))
	for _, c := range p.Status.Conditions {
		conditions = append(conditions, ConditionInfo{
			Type:    string(c.Type),
			Status:  string(c.Status),
			Reason:  c.Reason,
			Message: c.Message,
		})
	}

	created := ""
	if !p.CreationTimestamp.IsZero() {
		created = p.CreationTimestamp.UTC().Format(time.RFC3339)
	}

	status := string(p.Status.Phase)
	if p.DeletionTimestamp != nil {
		status = "Terminating"
	} else {
		for _, cs := range p.Status.ContainerStatuses {
			if cs.State.Waiting != nil && cs.State.Waiting.Reason != "" {
				status = cs.State.Waiting.Reason
				break
			}
		}
	}

	// Fetch events — best-effort, don't fail the whole request
	events := make([]EventInfo, 0)
	selector := fmt.Sprintf(
		"involvedObject.name=%s,involvedObject.namespace=%s,involvedObject.kind=Pod",
		name, namespace,
	)
	if evList, err := client.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{FieldSelector: selector}); err == nil {
		for _, ev := range evList.Items {
			events = append(events, EventInfo{
				Type:    ev.Type,
				Reason:  ev.Reason,
				Age:     podAge(ev.LastTimestamp.Time),
				From:    ev.Source.Component,
				Message: ev.Message,
				Count:   ev.Count,
			})
		}
	}

	owners := make([]RelatedResource, 0)
	for _, ref := range p.OwnerReferences {
		switch ref.Kind {
		case "ReplicaSet":
			if rs, err := client.AppsV1().ReplicaSets(namespace).Get(ctx, ref.Name, metav1.GetOptions{}); err == nil {
				for _, rsOwner := range rs.OwnerReferences {
					if rsOwner.Kind == "Deployment" {
						r := RelatedResource{Kind: "deployments", Name: rsOwner.Name}
						if d, err := client.AppsV1().Deployments(namespace).Get(ctx, rsOwner.Name, metav1.GetOptions{}); err == nil {
							var desired int32 = 1
							if d.Spec.Replicas != nil {
								desired = *d.Spec.Replicas
							}
							r.Age = podAge(d.CreationTimestamp.Time)
							r.CreatedAt = fmtTime(d.CreationTimestamp.Time)
							r.Extra = map[string]string{
								"ready":      fmt.Sprintf("%d/%d", d.Status.ReadyReplicas, desired),
								"up-to-date": fmt.Sprintf("%d", d.Status.UpdatedReplicas),
								"available":  fmt.Sprintf("%d", d.Status.AvailableReplicas),
							}
						}
						owners = append(owners, r)
					}
				}
			}
		case "StatefulSet":
			r := RelatedResource{Kind: "statefulsets", Name: ref.Name}
			if s, err := client.AppsV1().StatefulSets(namespace).Get(ctx, ref.Name, metav1.GetOptions{}); err == nil {
				var desired int32 = 1
				if s.Spec.Replicas != nil {
					desired = *s.Spec.Replicas
				}
				r.Age = podAge(s.CreationTimestamp.Time)
				r.CreatedAt = fmtTime(s.CreationTimestamp.Time)
				r.Extra = map[string]string{"ready": fmt.Sprintf("%d/%d", s.Status.ReadyReplicas, desired)}
			}
			owners = append(owners, r)
		case "DaemonSet":
			r := RelatedResource{Kind: "daemonsets", Name: ref.Name}
			if d, err := client.AppsV1().DaemonSets(namespace).Get(ctx, ref.Name, metav1.GetOptions{}); err == nil {
				r.Age = podAge(d.CreationTimestamp.Time)
				r.CreatedAt = fmtTime(d.CreationTimestamp.Time)
				r.Extra = map[string]string{
					"desired": fmt.Sprintf("%d", d.Status.DesiredNumberScheduled),
					"ready":   fmt.Sprintf("%d", d.Status.NumberReady),
				}
			}
			owners = append(owners, r)
		case "Job":
			if job, err := client.BatchV1().Jobs(namespace).Get(ctx, ref.Name, metav1.GetOptions{}); err == nil {
				owners = append(owners, jobToRelated(*job))
				for _, jobOwner := range job.OwnerReferences {
					if jobOwner.Kind == "CronJob" {
						r := RelatedResource{Kind: "cronjobs", Name: jobOwner.Name}
						if cj, err := client.BatchV1().CronJobs(namespace).Get(ctx, jobOwner.Name, metav1.GetOptions{}); err == nil {
							r.Age = podAge(cj.CreationTimestamp.Time)
							r.CreatedAt = fmtTime(cj.CreationTimestamp.Time)
							r.Extra = map[string]string{
								"schedule": cj.Spec.Schedule,
								"active":   fmt.Sprintf("%d", len(cj.Status.Active)),
							}
						}
						owners = append(owners, r)
					}
				}
			} else {
				owners = append(owners, RelatedResource{Kind: "jobs", Name: ref.Name})
			}
		}
	}

	return &PodDetail{
		Name:        p.Name,
		Namespace:   p.Namespace,
		Node:        p.Spec.NodeName,
		IP:          p.Status.PodIP,
		Status:      status,
		Phase:       string(p.Status.Phase),
		Age:         podAge(p.CreationTimestamp.Time),
		CreatedAt:   created,
		Labels:      p.Labels,
		Annotations: p.Annotations,
		Containers:  containers,
		Conditions:  conditions,
		Events:      events,
		Owners:      owners,
	}, nil
}

func podAge(t time.Time) string {
	if t.IsZero() {
		return "—"
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
}
