package kube

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ResourceSummary struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Age       string            `json:"age"`
	CreatedAt string            `json:"createdAt"`
	Status    string            `json:"status,omitempty"`
	Extra     map[string]string `json:"extra"`
}

type RelatedResource struct {
	Kind      string            `json:"kind"`
	Name      string            `json:"name"`
	Age       string            `json:"age,omitempty"`
	CreatedAt string            `json:"createdAt,omitempty"`
	Status    string            `json:"status,omitempty"`
	Extra     map[string]string `json:"extra,omitempty"`
}

type ResourceDetail struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Age         string            `json:"age"`
	CreatedAt   string            `json:"createdAt"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	Sections    []DetailSection   `json:"sections"`
	Related     []RelatedResource `json:"related,omitempty"`
}

type DetailSection struct {
	Title string `json:"title"`
	Items []KV   `json:"items"`
}

type KV struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func podToRelated(p corev1.Pod) RelatedResource {
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
	var restarts int32
	for _, cs := range p.Status.ContainerStatuses {
		if cs.Ready {
			ready++
		}
		restarts += cs.RestartCount
	}
	return RelatedResource{
		Kind: "pods", Name: p.Name,
		Age: podAge(p.CreationTimestamp.Time), CreatedAt: fmtTime(p.CreationTimestamp.Time),
		Status: status,
		Extra: map[string]string{
			"ready":    fmt.Sprintf("%d/%d", ready, len(p.Spec.Containers)),
			"restarts": fmt.Sprintf("%d", restarts),
		},
	}
}

func jobToRelated(j batchv1.Job) RelatedResource {
	status := "Running"
	for _, cond := range j.Status.Conditions {
		if string(cond.Type) == "Complete" && string(cond.Status) == "True" {
			status = "Complete"
			break
		}
		if string(cond.Type) == "Failed" && string(cond.Status) == "True" {
			status = "Failed"
			break
		}
	}
	completions := fmt.Sprintf("%d", j.Status.Succeeded)
	if j.Spec.Completions != nil {
		completions = fmt.Sprintf("%d/%d", j.Status.Succeeded, *j.Spec.Completions)
	}
	return RelatedResource{
		Kind: "jobs", Name: j.Name,
		Age: podAge(j.CreationTimestamp.Time), CreatedAt: fmtTime(j.CreationTimestamp.Time),
		Status: status,
		Extra:  map[string]string{"completions": completions},
	}
}

func labelSelectorStr(labels map[string]string) string {
	parts := make([]string, 0, len(labels))
	for k, v := range labels {
		parts = append(parts, k+"="+v)
	}
	sort.Strings(parts)
	return strings.Join(parts, ",")
}

func fmtTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

func labelsToString(m map[string]string) string {
	if len(m) == 0 {
		return ""
	}
	parts := make([]string, 0, len(m))
	for k, v := range m {
		parts = append(parts, k+"="+v)
	}
	sort.Strings(parts)
	return strings.Join(parts, ", ")
}

func ListResources(ctx context.Context, contextName, namespace, kind string) ([]ResourceSummary, error) {
	c, err := kubeClient(contextName)
	if err != nil {
		return nil, err
	}

	result := make([]ResourceSummary, 0)

	switch kind {
	case "deployments":
		list, err := c.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		for _, d := range list.Items {
			var desired int32 = 1
			if d.Spec.Replicas != nil {
				desired = *d.Spec.Replicas
			}
			result = append(result, ResourceSummary{
				Name: d.Name, Namespace: d.Namespace,
				Age: podAge(d.CreationTimestamp.Time), CreatedAt: fmtTime(d.CreationTimestamp.Time),
				Extra: map[string]string{
					"ready":      fmt.Sprintf("%d/%d", d.Status.ReadyReplicas, desired),
					"up-to-date": fmt.Sprintf("%d", d.Status.UpdatedReplicas),
					"available":  fmt.Sprintf("%d", d.Status.AvailableReplicas),
				},
			})
		}

	case "statefulsets":
		list, err := c.AppsV1().StatefulSets(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		for _, s := range list.Items {
			var desired int32 = 1
			if s.Spec.Replicas != nil {
				desired = *s.Spec.Replicas
			}
			result = append(result, ResourceSummary{
				Name: s.Name, Namespace: s.Namespace,
				Age: podAge(s.CreationTimestamp.Time), CreatedAt: fmtTime(s.CreationTimestamp.Time),
				Extra: map[string]string{
					"ready": fmt.Sprintf("%d/%d", s.Status.ReadyReplicas, desired),
				},
			})
		}

	case "daemonsets":
		list, err := c.AppsV1().DaemonSets(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		for _, d := range list.Items {
			result = append(result, ResourceSummary{
				Name: d.Name, Namespace: d.Namespace,
				Age: podAge(d.CreationTimestamp.Time), CreatedAt: fmtTime(d.CreationTimestamp.Time),
				Extra: map[string]string{
					"desired":    fmt.Sprintf("%d", d.Status.DesiredNumberScheduled),
					"current":    fmt.Sprintf("%d", d.Status.CurrentNumberScheduled),
					"ready":      fmt.Sprintf("%d", d.Status.NumberReady),
					"up-to-date": fmt.Sprintf("%d", d.Status.UpdatedNumberScheduled),
					"available":  fmt.Sprintf("%d", d.Status.NumberAvailable),
				},
			})
		}

	case "jobs":
		list, err := c.BatchV1().Jobs(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		for _, j := range list.Items {
			completions := fmt.Sprintf("%d", j.Status.Succeeded)
			if j.Spec.Completions != nil {
				completions = fmt.Sprintf("%d/%d", j.Status.Succeeded, *j.Spec.Completions)
			}
			status := "Running"
			for _, cond := range j.Status.Conditions {
				if string(cond.Type) == "Complete" && string(cond.Status) == "True" {
					status = "Complete"
					break
				}
				if string(cond.Type) == "Failed" && string(cond.Status) == "True" {
					status = "Failed"
					break
				}
			}
			result = append(result, ResourceSummary{
				Name: j.Name, Namespace: j.Namespace,
				Age: podAge(j.CreationTimestamp.Time), CreatedAt: fmtTime(j.CreationTimestamp.Time),
				Status: status,
				Extra:  map[string]string{"completions": completions},
			})
		}

	case "cronjobs":
		list, err := c.BatchV1().CronJobs(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		for _, cj := range list.Items {
			lastSchedule := "never"
			if cj.Status.LastScheduleTime != nil {
				lastSchedule = podAge(cj.Status.LastScheduleTime.Time)
			}
			suspend := "false"
			if cj.Spec.Suspend != nil && *cj.Spec.Suspend {
				suspend = "true"
			}
			result = append(result, ResourceSummary{
				Name: cj.Name, Namespace: cj.Namespace,
				Age: podAge(cj.CreationTimestamp.Time), CreatedAt: fmtTime(cj.CreationTimestamp.Time),
				Extra: map[string]string{
					"schedule":     cj.Spec.Schedule,
					"suspend":      suspend,
					"active":       fmt.Sprintf("%d", len(cj.Status.Active)),
					"lastSchedule": lastSchedule,
				},
			})
		}

	case "services":
		list, err := c.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		for _, svc := range list.Items {
			ports := make([]string, 0, len(svc.Spec.Ports))
			for _, p := range svc.Spec.Ports {
				if p.NodePort > 0 {
					ports = append(ports, fmt.Sprintf("%d:%d/%s", p.Port, p.NodePort, string(p.Protocol)))
				} else {
					ports = append(ports, fmt.Sprintf("%d/%s", p.Port, string(p.Protocol)))
				}
			}
			externalIP := "<none>"
			if len(svc.Status.LoadBalancer.Ingress) > 0 {
				ing := svc.Status.LoadBalancer.Ingress[0]
				if ing.IP != "" {
					externalIP = ing.IP
				} else {
					externalIP = ing.Hostname
				}
			} else if len(svc.Spec.ExternalIPs) > 0 {
				externalIP = strings.Join(svc.Spec.ExternalIPs, ",")
			}
			result = append(result, ResourceSummary{
				Name: svc.Name, Namespace: svc.Namespace,
				Age: podAge(svc.CreationTimestamp.Time), CreatedAt: fmtTime(svc.CreationTimestamp.Time),
				Extra: map[string]string{
					"type":       string(svc.Spec.Type),
					"clusterIP":  svc.Spec.ClusterIP,
					"externalIP": externalIP,
					"ports":      strings.Join(ports, ", "),
				},
			})
		}

	case "ingresses":
		list, err := c.NetworkingV1().Ingresses(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		for _, ing := range list.Items {
			hosts := make([]string, 0)
			for _, rule := range ing.Spec.Rules {
				hosts = append(hosts, rule.Host)
			}
			address := ""
			for _, lb := range ing.Status.LoadBalancer.Ingress {
				if lb.IP != "" {
					address = lb.IP
				} else {
					address = lb.Hostname
				}
				break
			}
			class := ""
			if ing.Spec.IngressClassName != nil {
				class = *ing.Spec.IngressClassName
			}
			result = append(result, ResourceSummary{
				Name: ing.Name, Namespace: ing.Namespace,
				Age: podAge(ing.CreationTimestamp.Time), CreatedAt: fmtTime(ing.CreationTimestamp.Time),
				Extra: map[string]string{
					"class":   class,
					"hosts":   strings.Join(hosts, ", "),
					"address": address,
				},
			})
		}

	case "configmaps":
		list, err := c.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		for _, cm := range list.Items {
			result = append(result, ResourceSummary{
				Name: cm.Name, Namespace: cm.Namespace,
				Age: podAge(cm.CreationTimestamp.Time), CreatedAt: fmtTime(cm.CreationTimestamp.Time),
				Extra: map[string]string{"data": fmt.Sprintf("%d", len(cm.Data))},
			})
		}

	case "secrets":
		list, err := c.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		for _, s := range list.Items {
			result = append(result, ResourceSummary{
				Name: s.Name, Namespace: s.Namespace,
				Age: podAge(s.CreationTimestamp.Time), CreatedAt: fmtTime(s.CreationTimestamp.Time),
				Extra: map[string]string{
					"type": string(s.Type),
					"data": fmt.Sprintf("%d", len(s.Data)),
				},
			})
		}

	case "persistentvolumeclaims":
		list, err := c.CoreV1().PersistentVolumeClaims(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		for _, pvc := range list.Items {
			capacity := ""
			if q, ok := pvc.Status.Capacity["storage"]; ok {
				capacity = q.String()
			}
			accessModes := make([]string, 0)
			for _, am := range pvc.Status.AccessModes {
				accessModes = append(accessModes, string(am))
			}
			sc := ""
			if pvc.Spec.StorageClassName != nil {
				sc = *pvc.Spec.StorageClassName
			}
			result = append(result, ResourceSummary{
				Name: pvc.Name, Namespace: pvc.Namespace,
				Age: podAge(pvc.CreationTimestamp.Time), CreatedAt: fmtTime(pvc.CreationTimestamp.Time),
				Status: string(pvc.Status.Phase),
				Extra: map[string]string{
					"volume":       pvc.Spec.VolumeName,
					"capacity":     capacity,
					"accessModes":  strings.Join(accessModes, ", "),
					"storageClass": sc,
				},
			})
		}

	default:
		return nil, fmt.Errorf("unknown resource kind: %s", kind)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result, nil
}

func GetResource(ctx context.Context, contextName, namespace, kind, name string) (*ResourceDetail, error) {
	c, err := kubeClient(contextName)
	if err != nil {
		return nil, err
	}

	switch kind {
	case "deployments":
		d, err := c.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		var desired int32 = 1
		if d.Spec.Replicas != nil {
			desired = *d.Spec.Replicas
		}
		paused := "false"
		if d.Spec.Paused {
			paused = "true"
		}
		progressDeadline := "600"
		if d.Spec.ProgressDeadlineSeconds != nil {
			progressDeadline = fmt.Sprintf("%d", *d.Spec.ProgressDeadlineSeconds)
		}
		revisionHistory := "10"
		if d.Spec.RevisionHistoryLimit != nil {
			revisionHistory = fmt.Sprintf("%d", *d.Spec.RevisionHistoryLimit)
		}
		detail := &ResourceDetail{
			Name: d.Name, Namespace: d.Namespace,
			Age: podAge(d.CreationTimestamp.Time), CreatedAt: fmtTime(d.CreationTimestamp.Time),
			Labels: d.Labels, Annotations: d.Annotations,
			Sections: []DetailSection{
				{Title: "Status", Items: []KV{
					{Key: "Ready", Value: fmt.Sprintf("%d/%d", d.Status.ReadyReplicas, desired)},
					{Key: "Up-to-date", Value: fmt.Sprintf("%d", d.Status.UpdatedReplicas)},
					{Key: "Available", Value: fmt.Sprintf("%d", d.Status.AvailableReplicas)},
					{Key: "Replicas", Value: fmt.Sprintf("%d", desired)},
				}},
				{Title: "Spec", Items: []KV{
					{Key: "Strategy", Value: string(d.Spec.Strategy.Type)},
					{Key: "Min Ready Seconds", Value: fmt.Sprintf("%d", d.Spec.MinReadySeconds)},
					{Key: "Progress Deadline", Value: progressDeadline + "s"},
					{Key: "Revision History Limit", Value: revisionHistory},
					{Key: "Paused", Value: paused},
				}},
			},
		}
		if d.Spec.Selector != nil && len(d.Spec.Selector.MatchLabels) > 0 {
			if pods, err := c.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: labelSelectorStr(d.Spec.Selector.MatchLabels)}); err == nil {
				for _, p := range pods.Items {
					detail.Related = append(detail.Related, podToRelated(p))
				}
			}
		}
		return detail, nil

	case "statefulsets":
		s, err := c.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		var desired int32 = 1
		if s.Spec.Replicas != nil {
			desired = *s.Spec.Replicas
		}
		detail := &ResourceDetail{
			Name: s.Name, Namespace: s.Namespace,
			Age: podAge(s.CreationTimestamp.Time), CreatedAt: fmtTime(s.CreationTimestamp.Time),
			Labels: s.Labels, Annotations: s.Annotations,
			Sections: []DetailSection{
				{Title: "Status", Items: []KV{
					{Key: "Ready", Value: fmt.Sprintf("%d/%d", s.Status.ReadyReplicas, desired)},
					{Key: "Current Replicas", Value: fmt.Sprintf("%d", s.Status.CurrentReplicas)},
					{Key: "Updated Replicas", Value: fmt.Sprintf("%d", s.Status.UpdatedReplicas)},
				}},
				{Title: "Spec", Items: []KV{
					{Key: "Service Name", Value: s.Spec.ServiceName},
					{Key: "Update Strategy", Value: string(s.Spec.UpdateStrategy.Type)},
					{Key: "Pod Management Policy", Value: string(s.Spec.PodManagementPolicy)},
					{Key: "Volume Claim Templates", Value: fmt.Sprintf("%d", len(s.Spec.VolumeClaimTemplates))},
				}},
			},
		}
		if s.Spec.Selector != nil && len(s.Spec.Selector.MatchLabels) > 0 {
			if pods, err := c.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: labelSelectorStr(s.Spec.Selector.MatchLabels)}); err == nil {
				for _, p := range pods.Items {
					detail.Related = append(detail.Related, podToRelated(p))
				}
			}
		}
		return detail, nil

	case "daemonsets":
		d, err := c.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		detail := &ResourceDetail{
			Name: d.Name, Namespace: d.Namespace,
			Age: podAge(d.CreationTimestamp.Time), CreatedAt: fmtTime(d.CreationTimestamp.Time),
			Labels: d.Labels, Annotations: d.Annotations,
			Sections: []DetailSection{
				{Title: "Status", Items: []KV{
					{Key: "Desired", Value: fmt.Sprintf("%d", d.Status.DesiredNumberScheduled)},
					{Key: "Current", Value: fmt.Sprintf("%d", d.Status.CurrentNumberScheduled)},
					{Key: "Ready", Value: fmt.Sprintf("%d", d.Status.NumberReady)},
					{Key: "Up-to-date", Value: fmt.Sprintf("%d", d.Status.UpdatedNumberScheduled)},
					{Key: "Available", Value: fmt.Sprintf("%d", d.Status.NumberAvailable)},
					{Key: "Misscheduled", Value: fmt.Sprintf("%d", d.Status.NumberMisscheduled)},
				}},
				{Title: "Spec", Items: []KV{
					{Key: "Update Strategy", Value: string(d.Spec.UpdateStrategy.Type)},
					{Key: "Min Ready Seconds", Value: fmt.Sprintf("%d", d.Spec.MinReadySeconds)},
					{Key: "Revision History Limit", Value: func() string {
						if d.Spec.RevisionHistoryLimit != nil {
							return fmt.Sprintf("%d", *d.Spec.RevisionHistoryLimit)
						}
						return "10"
					}()},
				}},
			},
		}
		if d.Spec.Selector != nil && len(d.Spec.Selector.MatchLabels) > 0 {
			if pods, err := c.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: labelSelectorStr(d.Spec.Selector.MatchLabels)}); err == nil {
				for _, p := range pods.Items {
					detail.Related = append(detail.Related, podToRelated(p))
				}
			}
		}
		return detail, nil

	case "jobs":
		j, err := c.BatchV1().Jobs(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		completions := fmt.Sprintf("%d", j.Status.Succeeded)
		if j.Spec.Completions != nil {
			completions = fmt.Sprintf("%d/%d", j.Status.Succeeded, *j.Spec.Completions)
		}
		detail := &ResourceDetail{
			Name: j.Name, Namespace: j.Namespace,
			Age: podAge(j.CreationTimestamp.Time), CreatedAt: fmtTime(j.CreationTimestamp.Time),
			Labels: j.Labels, Annotations: j.Annotations,
			Sections: func() []DetailSection {
				statusItems := []KV{
					{Key: "Completions", Value: completions},
					{Key: "Active", Value: fmt.Sprintf("%d", j.Status.Active)},
					{Key: "Failed", Value: fmt.Sprintf("%d", j.Status.Failed)},
				}
				if j.Status.StartTime != nil {
					statusItems = append(statusItems, KV{Key: "Start Time", Value: j.Status.StartTime.UTC().Format("2006-01-02 15:04:05 UTC")})
				}
				if j.Status.CompletionTime != nil {
					statusItems = append(statusItems, KV{Key: "Completion Time", Value: j.Status.CompletionTime.UTC().Format("2006-01-02 15:04:05 UTC")})
				}
				if j.Status.StartTime != nil && j.Status.CompletionTime != nil {
					dur := j.Status.CompletionTime.Sub(j.Status.StartTime.Time)
					statusItems = append(statusItems, KV{Key: "Duration", Value: dur.Round(time.Second).String()})
				}
				specItems := []KV{}
				if j.Spec.Parallelism != nil {
					specItems = append(specItems, KV{Key: "Parallelism", Value: fmt.Sprintf("%d", *j.Spec.Parallelism)})
				}
				if j.Spec.BackoffLimit != nil {
					specItems = append(specItems, KV{Key: "Backoff Limit", Value: fmt.Sprintf("%d", *j.Spec.BackoffLimit)})
				}
				if j.Spec.ActiveDeadlineSeconds != nil {
					specItems = append(specItems, KV{Key: "Active Deadline", Value: fmt.Sprintf("%ds", *j.Spec.ActiveDeadlineSeconds)})
				}
				sections := []DetailSection{{Title: "Status", Items: statusItems}}
				if len(specItems) > 0 {
					sections = append(sections, DetailSection{Title: "Spec", Items: specItems})
				}
				return sections
			}(),
		}
		if pods, err := c.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: "job-name=" + name}); err == nil {
			for _, p := range pods.Items {
				detail.Related = append(detail.Related, podToRelated(p))
			}
		}
		return detail, nil

	case "cronjobs":
		cj, err := c.BatchV1().CronJobs(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		lastSchedule := "never"
		if cj.Status.LastScheduleTime != nil {
			lastSchedule = cj.Status.LastScheduleTime.UTC().Format("2006-01-02 15:04:05 UTC")
		}
		suspend := "false"
		if cj.Spec.Suspend != nil && *cj.Spec.Suspend {
			suspend = "true"
		}
		detail := &ResourceDetail{
			Name: cj.Name, Namespace: cj.Namespace,
			Age: podAge(cj.CreationTimestamp.Time), CreatedAt: fmtTime(cj.CreationTimestamp.Time),
			Labels: cj.Labels, Annotations: cj.Annotations,
			Sections: []DetailSection{
				{Title: "Spec", Items: func() []KV {
					items := []KV{
						{Key: "Schedule", Value: cj.Spec.Schedule},
						{Key: "Suspend", Value: suspend},
						{Key: "Concurrency Policy", Value: string(cj.Spec.ConcurrencyPolicy)},
					}
					if cj.Spec.StartingDeadlineSeconds != nil {
						items = append(items, KV{Key: "Starting Deadline", Value: fmt.Sprintf("%ds", *cj.Spec.StartingDeadlineSeconds)})
					}
					items = append(items,
						KV{Key: "Successful Jobs History", Value: func() string {
							if cj.Spec.SuccessfulJobsHistoryLimit != nil {
								return fmt.Sprintf("%d", *cj.Spec.SuccessfulJobsHistoryLimit)
							}
							return "3"
						}()},
						KV{Key: "Failed Jobs History", Value: func() string {
							if cj.Spec.FailedJobsHistoryLimit != nil {
								return fmt.Sprintf("%d", *cj.Spec.FailedJobsHistoryLimit)
							}
							return "1"
						}()},
					)
					return items
				}()},
				{Title: "Status", Items: func() []KV {
					items := []KV{
						{Key: "Active", Value: fmt.Sprintf("%d", len(cj.Status.Active))},
						{Key: "Last Schedule", Value: lastSchedule},
					}
					if cj.Status.LastSuccessfulTime != nil {
						items = append(items, KV{Key: "Last Successful", Value: cj.Status.LastSuccessfulTime.UTC().Format("2006-01-02 15:04:05 UTC")})
					}
					return items
				}()},
			},
		}
		// list jobs owned by this cronjob
		if jobList, err := c.BatchV1().Jobs(namespace).List(ctx, metav1.ListOptions{}); err == nil {
			for _, j := range jobList.Items {
				for _, ref := range j.OwnerReferences {
					if ref.Kind == "CronJob" && ref.Name == cj.Name {
						detail.Related = append(detail.Related, jobToRelated(j))
						break
					}
				}
			}
		}
		return detail, nil

	case "services":
		svc, err := c.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		ports := make([]string, 0, len(svc.Spec.Ports))
		for _, p := range svc.Spec.Ports {
			entry := fmt.Sprintf("%d/%s", p.Port, string(p.Protocol))
			if p.Name != "" {
				entry = p.Name + ": " + entry
			}
			if p.NodePort > 0 {
				entry += fmt.Sprintf(" → %d", p.NodePort)
			}
			ports = append(ports, entry)
		}
		externalIP := "<none>"
		if len(svc.Status.LoadBalancer.Ingress) > 0 {
			lb := svc.Status.LoadBalancer.Ingress[0]
			if lb.IP != "" {
				externalIP = lb.IP
			} else {
				externalIP = lb.Hostname
			}
		} else if len(svc.Spec.ExternalIPs) > 0 {
			externalIP = strings.Join(svc.Spec.ExternalIPs, ", ")
		}
		specItems := []KV{
			{Key: "Type", Value: string(svc.Spec.Type)},
			{Key: "Cluster IP", Value: svc.Spec.ClusterIP},
			{Key: "External IP", Value: externalIP},
			{Key: "Ports", Value: strings.Join(ports, "; ")},
			{Key: "Session Affinity", Value: string(svc.Spec.SessionAffinity)},
			{Key: "Selector", Value: labelsToString(svc.Spec.Selector)},
		}
		if svc.Spec.Type == "LoadBalancer" || svc.Spec.Type == "NodePort" {
			specItems = append(specItems, KV{Key: "External Traffic Policy", Value: string(svc.Spec.ExternalTrafficPolicy)})
		}
		sections := []DetailSection{{Title: "Spec", Items: specItems}}

		// endpoints
		if eps, err := c.CoreV1().Endpoints(namespace).Get(ctx, name, metav1.GetOptions{}); err == nil {
			epItems := make([]KV, 0)
			for _, subset := range eps.Subsets {
				portNames := make([]string, 0, len(subset.Ports))
				for _, p := range subset.Ports {
					portNames = append(portNames, fmt.Sprintf("%d/%s", p.Port, string(p.Protocol)))
				}
				portsStr := strings.Join(portNames, ", ")
				for _, addr := range subset.Addresses {
					target := addr.IP
					if addr.TargetRef != nil {
						target = addr.TargetRef.Name + " (" + addr.IP + ")"
					}
					epItems = append(epItems, KV{Key: target, Value: portsStr})
				}
				for _, addr := range subset.NotReadyAddresses {
					target := addr.IP + " (not ready)"
					if addr.TargetRef != nil {
						target = addr.TargetRef.Name + " (not ready, " + addr.IP + ")"
					}
					epItems = append(epItems, KV{Key: target, Value: portsStr})
				}
			}
			if len(epItems) > 0 {
				sections = append(sections, DetailSection{Title: "Endpoints", Items: epItems})
			}
		}

		detail := &ResourceDetail{
			Name: svc.Name, Namespace: svc.Namespace,
			Age: podAge(svc.CreationTimestamp.Time), CreatedAt: fmtTime(svc.CreationTimestamp.Time),
			Labels: svc.Labels, Annotations: svc.Annotations,
			Sections: sections,
		}
		// pods matching selector
		if len(svc.Spec.Selector) > 0 {
			if pods, err := c.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: labelSelectorStr(svc.Spec.Selector)}); err == nil {
				for _, p := range pods.Items {
					detail.Related = append(detail.Related, podToRelated(p))
				}
			}
		}
		return detail, nil

	case "ingresses":
		ing, err := c.NetworkingV1().Ingresses(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		class := ""
		if ing.Spec.IngressClassName != nil {
			class = *ing.Spec.IngressClassName
		}
		address := ""
		for _, lb := range ing.Status.LoadBalancer.Ingress {
			if lb.IP != "" {
				address = lb.IP
			} else {
				address = lb.Hostname
			}
			break
		}
		specItems := []KV{{Key: "Class", Value: class}}
		if address != "" {
			specItems = append(specItems, KV{Key: "Address", Value: address})
		}
		if ing.Spec.DefaultBackend != nil && ing.Spec.DefaultBackend.Service != nil {
			svc := ing.Spec.DefaultBackend.Service
			port := fmt.Sprintf("%d", svc.Port.Number)
			if svc.Port.Name != "" {
				port = svc.Port.Name
			}
			specItems = append(specItems, KV{Key: "Default Backend", Value: svc.Name + ":" + port})
		}
		sections := []DetailSection{{Title: "Spec", Items: specItems}}

		// rules: one KV per path
		if len(ing.Spec.Rules) > 0 {
			ruleItems := make([]KV, 0)
			for _, r := range ing.Spec.Rules {
				host := r.Host
				if host == "" {
					host = "*"
				}
				if r.HTTP != nil {
					for _, p := range r.HTTP.Paths {
						svcName := ""
						svcPort := ""
						if p.Backend.Service != nil {
							svcName = p.Backend.Service.Name
							if p.Backend.Service.Port.Name != "" {
								svcPort = p.Backend.Service.Port.Name
							} else {
								svcPort = fmt.Sprintf("%d", p.Backend.Service.Port.Number)
							}
						}
						pathType := ""
						if p.PathType != nil {
							pathType = string(*p.PathType)
						}
						key := host + p.Path
						if pathType != "" {
							key += " (" + pathType + ")"
						}
						ruleItems = append(ruleItems, KV{Key: key, Value: svcName + ":" + svcPort})
					}
				} else {
					ruleItems = append(ruleItems, KV{Key: host, Value: ""})
				}
			}
			sections = append(sections, DetailSection{Title: "Rules", Items: ruleItems})
		}

		// TLS
		if len(ing.Spec.TLS) > 0 {
			tlsItems := make([]KV, 0, len(ing.Spec.TLS))
			for _, t := range ing.Spec.TLS {
				tlsItems = append(tlsItems, KV{Key: strings.Join(t.Hosts, ", "), Value: t.SecretName})
			}
			sections = append(sections, DetailSection{Title: "TLS", Items: tlsItems})
		}

		// collect unique service names referenced by this ingress
		svcNames := make(map[string]struct{})
		if ing.Spec.DefaultBackend != nil && ing.Spec.DefaultBackend.Service != nil {
			svcNames[ing.Spec.DefaultBackend.Service.Name] = struct{}{}
		}
		for _, r := range ing.Spec.Rules {
			if r.HTTP != nil {
				for _, p := range r.HTTP.Paths {
					if p.Backend.Service != nil {
						svcNames[p.Backend.Service.Name] = struct{}{}
					}
				}
			}
		}
		detail := &ResourceDetail{
			Name: ing.Name, Namespace: ing.Namespace,
			Age: podAge(ing.CreationTimestamp.Time), CreatedAt: fmtTime(ing.CreationTimestamp.Time),
			Labels: ing.Labels, Annotations: ing.Annotations,
			Sections: sections,
		}
		for svcName := range svcNames {
			r := RelatedResource{Kind: "services", Name: svcName}
			if svc, err := c.CoreV1().Services(namespace).Get(ctx, svcName, metav1.GetOptions{}); err == nil {
				ports := make([]string, 0, len(svc.Spec.Ports))
				for _, p := range svc.Spec.Ports {
					ports = append(ports, fmt.Sprintf("%d/%s", p.Port, string(p.Protocol)))
				}
				r.Age = podAge(svc.CreationTimestamp.Time)
				r.CreatedAt = fmtTime(svc.CreationTimestamp.Time)
				r.Extra = map[string]string{
					"type":      string(svc.Spec.Type),
					"clusterIP": svc.Spec.ClusterIP,
					"ports":     strings.Join(ports, ", "),
				}
			}
			detail.Related = append(detail.Related, r)
		}
		return detail, nil

	case "configmaps":
		cm, err := c.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		keys := make([]string, 0, len(cm.Data))
		for k := range cm.Data {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		items := make([]KV, 0, len(keys))
		for _, k := range keys {
			v := cm.Data[k]
			if len(v) > 200 {
				v = v[:200] + "…"
			}
			items = append(items, KV{Key: k, Value: v})
		}
		return &ResourceDetail{
			Name: cm.Name, Namespace: cm.Namespace,
			Age: podAge(cm.CreationTimestamp.Time), CreatedAt: fmtTime(cm.CreationTimestamp.Time),
			Labels: cm.Labels, Annotations: cm.Annotations,
			Sections: []DetailSection{
				{Title: "Data", Items: items},
			},
		}, nil

	case "secrets":
		s, err := c.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		keys := make([]string, 0, len(s.Data))
		for k := range s.Data {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		items := make([]KV, 0, len(keys))
		for _, k := range keys {
			items = append(items, KV{Key: k, Value: fmt.Sprintf("<redacted> (%d bytes)", len(s.Data[k]))})
		}
		return &ResourceDetail{
			Name: s.Name, Namespace: s.Namespace,
			Age: podAge(s.CreationTimestamp.Time), CreatedAt: fmtTime(s.CreationTimestamp.Time),
			Labels: s.Labels, Annotations: s.Annotations,
			Sections: []DetailSection{
				{Title: "Data", Items: items},
			},
		}, nil

	case "persistentvolumeclaims":
		pvc, err := c.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		capacity := ""
		if q, ok := pvc.Status.Capacity["storage"]; ok {
			capacity = q.String()
		}
		accessModes := make([]string, 0)
		for _, am := range pvc.Status.AccessModes {
			accessModes = append(accessModes, string(am))
		}
		sc := ""
		if pvc.Spec.StorageClassName != nil {
			sc = *pvc.Spec.StorageClassName
		}
		requestedStorage := ""
		if q, ok := pvc.Spec.Resources.Requests["storage"]; ok {
			requestedStorage = q.String()
		}
		sections := []DetailSection{
			{Title: "Status", Items: []KV{
				{Key: "Phase", Value: string(pvc.Status.Phase)},
				{Key: "Volume", Value: pvc.Spec.VolumeName},
				{Key: "Capacity", Value: capacity},
				{Key: "Requested", Value: requestedStorage},
				{Key: "Access Modes", Value: strings.Join(accessModes, ", ")},
				{Key: "Storage Class", Value: sc},
				{Key: "Volume Mode", Value: func() string {
					if pvc.Spec.VolumeMode != nil {
						return string(*pvc.Spec.VolumeMode)
					}
					return "Filesystem"
				}()},
			}},
		}
		// bound PV details
		if pvc.Spec.VolumeName != "" {
			if pv, err := c.CoreV1().PersistentVolumes().Get(ctx, pvc.Spec.VolumeName, metav1.GetOptions{}); err == nil {
				pvItems := []KV{
					{Key: "Reclaim Policy", Value: string(pv.Spec.PersistentVolumeReclaimPolicy)},
					{Key: "Status", Value: string(pv.Status.Phase)},
				}
				if pv.Spec.StorageClassName != "" {
					pvItems = append(pvItems, KV{Key: "Storage Class", Value: pv.Spec.StorageClassName})
				}
				sections = append(sections, DetailSection{Title: "Persistent Volume", Items: pvItems})
			}
		}
		detail := &ResourceDetail{
			Name: pvc.Name, Namespace: pvc.Namespace,
			Age: podAge(pvc.CreationTimestamp.Time), CreatedAt: fmtTime(pvc.CreationTimestamp.Time),
			Labels: pvc.Labels, Annotations: pvc.Annotations,
			Sections: sections,
		}
		// pods using this PVC
		if pods, err := c.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{}); err == nil {
			for _, p := range pods.Items {
				for _, vol := range p.Spec.Volumes {
					if vol.PersistentVolumeClaim != nil && vol.PersistentVolumeClaim.ClaimName == name {
						detail.Related = append(detail.Related, podToRelated(p))
						break
					}
				}
			}
		}
		return detail, nil

	default:
		return nil, fmt.Errorf("unknown resource kind: %s", kind)
	}
}
