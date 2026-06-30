package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	store := NewStore(filepath.Join(t.TempDir(), dataFileName))
	now := time.Date(2026, 6, 30, 10, 0, 0, 0, time.FixedZone("CST", 8*60*60))
	store.now = func() time.Time { return now }
	if err := store.Load(); err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	return store
}

func TestTaskLifecyclePersists(t *testing.T) {
	store := newTestStore(t)

	task, err := store.Add(" Buy milk ")
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	if task.ID != 1 || task.Title != "Buy milk" || task.Status != statusPending {
		t.Fatalf("unexpected task: %+v", task)
	}

	if err := store.Done(task.ID); err != nil {
		t.Fatalf("Done() error = %v", err)
	}
	if err := store.Undo(task.ID); err != nil {
		t.Fatalf("Undo() error = %v", err)
	}
	if err := store.Edit(task.ID, "Buy bread"); err != nil {
		t.Fatalf("Edit() error = %v", err)
	}

	reloaded := NewStore(store.path)
	if err := reloaded.Load(); err != nil {
		t.Fatalf("reloaded Load() error = %v", err)
	}
	tasks, err := reloaded.List("all")
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(tasks) != 1 || tasks[0].Title != "Buy bread" || tasks[0].Status != statusPending {
		t.Fatalf("unexpected reloaded tasks: %+v", tasks)
	}

	if err := reloaded.Delete(task.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	tasks, err = reloaded.List("all")
	if err != nil {
		t.Fatalf("List() after delete error = %v", err)
	}
	if len(tasks) != 0 {
		t.Fatalf("deleted task still listed: %+v", tasks)
	}
}

func TestListSortsPendingBeforeDoneThenID(t *testing.T) {
	store := newTestStore(t)
	first, _ := store.Add("first")
	second, _ := store.Add("second")
	third, _ := store.Add("third")
	if err := store.Done(first.ID); err != nil {
		t.Fatalf("Done(first) error = %v", err)
	}
	if err := store.Done(third.ID); err != nil {
		t.Fatalf("Done(third) error = %v", err)
	}

	tasks, err := store.List("all")
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	gotIDs := []int{tasks[0].ID, tasks[1].ID, tasks[2].ID}
	wantIDs := []int{second.ID, first.ID, third.ID}
	for i := range wantIDs {
		if gotIDs[i] != wantIDs[i] {
			t.Fatalf("got IDs %v, want %v", gotIDs, wantIDs)
		}
	}
}

func TestValidationErrors(t *testing.T) {
	store := newTestStore(t)

	if _, err := store.Add("   "); err == nil {
		t.Fatal("Add(empty) error = nil")
	}
	if _, err := store.Add(strings.Repeat("中", 201)); err == nil {
		t.Fatal("Add(long title) error = nil")
	}
	if _, err := store.Add("line\nbreak"); err == nil {
		t.Fatal("Add(control char title) error = nil")
	}
	if err := store.Done(99); err == nil {
		t.Fatal("Done(missing) error = nil")
	}
	if _, err := store.List("invalid"); err == nil {
		t.Fatal("List(invalid) error = nil")
	}
}

func TestLoadCorruptedJSONReturnsClearError(t *testing.T) {
	path := filepath.Join(t.TempDir(), dataFileName)
	if err := os.WriteFile(path, []byte("{bad json"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	store := NewStore(path)
	err := store.Load()
	if err == nil {
		t.Fatal("Load(corrupted) error = nil")
	}
	if !strings.Contains(err.Error(), "数据文件解析失败") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWriteFailureDoesNotMutateMemoryState(t *testing.T) {
	base := t.TempDir()
	blocker := filepath.Join(base, "blocker")
	if err := os.WriteFile(blocker, []byte("not a directory"), 0644); err != nil {
		t.Fatalf("WriteFile(blocker) error = %v", err)
	}

	store := NewStore(filepath.Join(blocker, dataFileName))
	now := time.Date(2026, 6, 30, 10, 0, 0, 0, time.UTC)
	store.now = func() time.Time { return now }
	store.data = DataFile{
		Version: dataVersion,
		NextID:  2,
		Tasks: []Task{
			{ID: 1, Title: "first", Status: statusPending, CreatedAt: now, UpdatedAt: now},
		},
	}

	if _, err := store.Add("second"); err == nil {
		t.Fatal("Add() error = nil")
	}
	if store.data.NextID != 2 || len(store.data.Tasks) != 1 {
		t.Fatalf("Add mutated memory after save failure: %+v", store.data)
	}

	if err := store.Done(1); err == nil {
		t.Fatal("Done() error = nil")
	}
	if store.data.Tasks[0].Status != statusPending {
		t.Fatalf("Done mutated memory after save failure: %+v", store.data.Tasks[0])
	}

	if err := store.Edit(1, "changed"); err == nil {
		t.Fatal("Edit() error = nil")
	}
	if store.data.Tasks[0].Title != "first" {
		t.Fatalf("Edit mutated memory after save failure: %+v", store.data.Tasks[0])
	}

	if err := store.Delete(1); err == nil {
		t.Fatal("Delete() error = nil")
	}
	if len(store.data.Tasks) != 1 || store.data.Tasks[0].ID != 1 {
		t.Fatalf("Delete mutated memory after save failure: %+v", store.data)
	}
}
