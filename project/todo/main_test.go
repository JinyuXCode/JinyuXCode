package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunAddListAndExitCodes(t *testing.T) {
	dir := t.TempDir()
	dataPath := filepath.Join(dir, dataFileName)

	var stdout, stderr bytes.Buffer
	if code := runWithPath([]string{"add", "Buy", "milk"}, dataPath, &stdout, &stderr); code != 0 {
		t.Fatalf("add code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "已新增任务: 1") {
		t.Fatalf("unexpected add output: %q", stdout.String())
	}
	if _, err := os.Stat(dataPath); err != nil {
		t.Fatalf("todo.json was not created: %v", err)
	}

	stdout.Reset()
	stderr.Reset()
	if code := runWithPath([]string{"list", "--status", "pending"}, dataPath, &stdout, &stderr); code != 0 {
		t.Fatalf("list code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Buy milk") || !strings.Contains(stdout.String(), "pending") {
		t.Fatalf("unexpected list output: %q", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	if code := runWithPath([]string{"done", "999"}, dataPath, &stdout, &stderr); code != 2 {
		t.Fatalf("done missing code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "任务不存在") {
		t.Fatalf("unexpected error output: %q", stderr.String())
	}
}

func TestArgumentErrorsDoNotLoadCorruptedDataFile(t *testing.T) {
	dataPath := filepath.Join(t.TempDir(), dataFileName)
	if err := os.WriteFile(dataPath, []byte("{bad json"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	tests := []struct {
		name       string
		args       []string
		wantOutput string
	}{
		{name: "unknown command", args: []string{"unknown"}, wantOutput: "未知命令"},
		{name: "missing title", args: []string{"add"}, wantOutput: "缺少任务标题"},
		{name: "invalid id", args: []string{"done", "abc"}, wantOutput: "id 必须是正整数"},
		{name: "invalid status", args: []string{"list", "--status", "bad"}, wantOutput: "status 仅支持"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if code := runWithPath(tt.args, dataPath, &stdout, &stderr); code != 2 {
				t.Fatalf("code = %d, stderr = %s", code, stderr.String())
			}
			if strings.Contains(stderr.String(), "数据文件解析失败") {
				t.Fatalf("argument error was masked by data load error: %q", stderr.String())
			}
			if !strings.Contains(stderr.String(), tt.wantOutput) {
				t.Fatalf("stderr = %q, want contains %q", stderr.String(), tt.wantOutput)
			}
		})
	}
}
