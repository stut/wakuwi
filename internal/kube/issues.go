package kube

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Issue struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Age       string `json:"age"`
	Severity  string `json:"severity"` // "error" | "warning"
	Message   string `json:"message"`
	Reason    string `json:"reason"` // normalized category for filtering
}

func ListIssues(ctx context.Context, contextName string) ([]Issue, error) {
	c, err := kubeClient(contextName)
	if err != nil {
		return nil, err
	}

	var issues []Issue

	// pods
	pods, err := c.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err == nil {
		for _, p := range pods.Items {
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
			switch status {
			case "Running", "Succeeded", "Completed":
				// check restarts
				var restarts int32
				for _, cs := range p.Status.ContainerStatuses {
					restarts += cs.RestartCount
				}
				if restarts >= 10 {
					issues = append(issues, Issue{
						Kind: "pods", Name: p.Name, Namespace: p.Namespace,
						Age: podAge(p.CreationTimestamp.Time), Severity: "warning",
						Message: fmt.Sprintf("%d restarts", restarts), Reason: "High Restarts",
					})
				}
			case "Pending":
				issues = append(issues, Issue{
					Kind: "pods", Name: p.Name, Namespace: p.Namespace,
					Age: podAge(p.CreationTimestamp.Time), Severity: "warning",
					Message: "Pending", Reason: "Pending",
				})
			default:
				issues = append(issues, Issue{
					Kind: "pods", Name: p.Name, Namespace: p.Namespace,
					Age: podAge(p.CreationTimestamp.Time), Severity: "error",
					Message: status, Reason: status,
				})
			}
		}
	}

	// deployments
	deployments, err := c.AppsV1().Deployments("").List(ctx, metav1.ListOptions{})
	if err == nil {
		for _, d := range deployments.Items {
			var desired int32 = 1
			if d.Spec.Replicas != nil {
				desired = *d.Spec.Replicas
			}
			if desired == 0 {
				continue
			}
			if d.Status.ReadyReplicas == 0 {
				issues = append(issues, Issue{
					Kind: "deployments", Name: d.Name, Namespace: d.Namespace,
					Age: podAge(d.CreationTimestamp.Time), Severity: "error",
					Message: fmt.Sprintf("0/%d replicas ready", desired), Reason: "No Replicas Ready",
				})
			} else if d.Status.UnavailableReplicas > 0 {
				issues = append(issues, Issue{
					Kind: "deployments", Name: d.Name, Namespace: d.Namespace,
					Age: podAge(d.CreationTimestamp.Time), Severity: "warning",
					Message: fmt.Sprintf("%d/%d replicas ready", d.Status.ReadyReplicas, desired), Reason: "Degraded",
				})
			}
		}
	}

	// jobs
	jobs, err := c.BatchV1().Jobs("").List(ctx, metav1.ListOptions{})
	if err == nil {
		for _, j := range jobs.Items {
			if j.Status.Failed > 0 {
				sev := "warning"
				msg := fmt.Sprintf("%d failed", j.Status.Failed)
				reason := "Job Failed"
				if j.Spec.BackoffLimit != nil && j.Status.Failed > *j.Spec.BackoffLimit {
					sev = "error"
					msg = fmt.Sprintf("failed after %d attempts", j.Status.Failed)
				}
				issues = append(issues, Issue{
					Kind: "jobs", Name: j.Name, Namespace: j.Namespace,
					Age: podAge(j.CreationTimestamp.Time), Severity: sev,
					Message: msg, Reason: reason,
				})
			}
		}
	}

	// pvcs
	pvcs, err := c.CoreV1().PersistentVolumeClaims("").List(ctx, metav1.ListOptions{})
	if err == nil {
		for _, pvc := range pvcs.Items {
			if pvc.Status.Phase == "Pending" {
				issues = append(issues, Issue{
					Kind: "persistentvolumeclaims", Name: pvc.Name, Namespace: pvc.Namespace,
					Age: podAge(pvc.CreationTimestamp.Time), Severity: "warning",
					Message: "Pending (unbound)", Reason: "PVC Unbound",
				})
			}
		}
	}

	return issues, nil
}
