package kubernetes

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/lihaiya/aiops/pkg/aiops/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Execute list/get/describe/logs；禁止 delete/apply/patch/create。
func Execute(plan types.ActionPlan) (types.ExecuteResponse, error) {
	switch plan.Action {
	case "delete", "apply", "patch", "create":
		return types.ExecuteResponse{}, fmt.Errorf("action %s forbidden", plan.Action)
	case "get", "list", "describe", "logs":
		client, err := newClientset()
		if err != nil {
			return types.ExecuteResponse{}, err
		}
		ctx := context.Background()
		ns := paramString(plan.Parameters, "namespace", "default")
		switch plan.Action {
		case "list":
			resource := paramString(plan.Parameters, "resource", "pods")
			if resource == "pods" {
				list, err := client.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{})
				if err != nil {
					return types.ExecuteResponse{}, err
				}
				names := make([]string, 0, len(list.Items))
				for _, p := range list.Items {
					names = append(names, p.Name)
				}
				return types.ExecuteResponse{Status: "ok", Output: map[string]interface{}{"items": names}}, nil
			}
			return types.ExecuteResponse{}, fmt.Errorf("unsupported resource: %s", resource)
		case "get", "describe":
			name := paramString(plan.Parameters, "name", "")
			if name == "" {
				return types.ExecuteResponse{}, fmt.Errorf("name required")
			}
			pod, err := client.CoreV1().Pods(ns).Get(ctx, name, metav1.GetOptions{})
			if err != nil {
				return types.ExecuteResponse{}, err
			}
			return types.ExecuteResponse{Status: "ok", Output: pod}, nil
		case "logs":
			name := paramString(plan.Parameters, "name", "")
			if name == "" {
				return types.ExecuteResponse{}, fmt.Errorf("name required")
			}
			tail := int64(200)
			if t := paramString(plan.Parameters, "tailLines", ""); t != "" {
				if v, err := strconv.ParseInt(t, 10, 64); err == nil && v > 0 && v <= 200 {
					tail = v
				}
			}
			req := client.CoreV1().Pods(ns).GetLogs(name, &corev1.PodLogOptions{TailLines: &tail})
			stream, err := req.Stream(ctx)
			if err != nil {
				return types.ExecuteResponse{}, err
			}
			defer stream.Close()
			b, _ := io.ReadAll(stream)
			return types.ExecuteResponse{Status: "ok", Output: map[string]interface{}{"logs": string(b)}}, nil
		}
	default:
		return types.ExecuteResponse{}, fmt.Errorf("unsupported action: %s", plan.Action)
	}
	return types.ExecuteResponse{}, fmt.Errorf("unreachable")
}

func newClientset() (*kubernetes.Clientset, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		kc := os.Getenv("KUBECONFIG")
		if kc == "" {
			home, _ := os.UserHomeDir()
			kc = home + "/.kube/config"
		}
		cfg, err = clientcmd.BuildConfigFromFlags("", kc)
		if err != nil {
			return nil, fmt.Errorf("k8s config: %w", err)
		}
	}
	return kubernetes.NewForConfig(cfg)
}

func paramString(m map[string]interface{}, key, def string) string {
	if m == nil {
		return def
	}
	v, ok := m[key]
	if !ok {
		return def
	}
	switch t := v.(type) {
	case string:
		if t != "" {
			return t
		}
	}
	return def
}

// ForbiddenAction 是否禁止的动作。
func ForbiddenAction(action string) bool {
	switch strings.ToLower(action) {
	case "delete", "apply", "patch", "create":
		return true
	}
	return false
}
