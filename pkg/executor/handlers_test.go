package executor

import (
	"os"
	"strings"
	"testing"
)

// TestExecutorNoPluginImport 确保 executor 包不直接 import 具体插件实现。
func TestExecutorNoPluginImport(t *testing.T) {
	src := readSource(t, "handlers.go")
	if strings.Contains(src, "pkg/plugins/kubernetes") || strings.Contains(src, "pkg/plugins/mock") {
		t.Fatal("executor must not import concrete plugins")
	}
	if !strings.Contains(src, "pkg/plugins") {
		t.Fatal("executor should use pkg/plugins client only")
	}
}

func TestConfirmMustReEvaluate(t *testing.T) {
	src := readSource(t, "handlers.go")
	confirm := src[strings.Index(src, "func (h *Handler) Confirm"):strings.Index(src, "func (h *Handler) audit")]
	if strings.Count(confirm, "/internal/v1/evaluate") < 1 {
		t.Fatal("confirm must call policy evaluate before execute")
	}
}

func readSource(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(name)
	if err != nil {
		t.Skip(err)
	}
	return string(b)
}
