package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const (
	dataVersion   = 1
	statusPending = "pending"
	statusDone    = "done"
	maxTitleRunes = 200
)

var (
	errTitleEmpty    = errors.New("标题不能为空")
	errTitleTooLong  = fmt.Errorf("标题不能超过 %d 个字符", maxTitleRunes)
	errTitleControl  = errors.New("标题不能包含控制字符")
	errTaskNotFound  = errors.New("任务不存在")
	errInvalidStatus = errors.New("status 仅支持 pending、done 或 all")
)

type Task struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type DataFile struct {
	Version int    `json:"version"`
	NextID  int    `json:"nextId"`
	Tasks   []Task `json:"tasks"`
}

type Store struct {
	path string
	now  func() time.Time
	data DataFile
}

func NewStore(path string) *Store {
	return &Store{
		path: path,
		now:  time.Now,
	}
}

func (s *Store) Load() error {
	content, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			data := emptyData()
			if err := s.saveData(data); err != nil {
				return err
			}
			s.data = data
			return nil
		}
		return fmt.Errorf("数据文件不可读: %w", err)
	}

	var data DataFile
	if err := json.Unmarshal(content, &data); err != nil {
		return fmt.Errorf("数据文件解析失败，请备份或重建 %s: %w", filepath.Base(s.path), err)
	}

	if err := validateData(data); err != nil {
		return fmt.Errorf("数据文件内容无效，请备份或重建 %s: %w", filepath.Base(s.path), err)
	}

	sortTasksByID(data.Tasks)
	s.data = data
	return nil
}

func (s *Store) Save() error {
	data := cloneData(s.data)
	return s.saveData(data)
}

func (s *Store) saveData(data DataFile) error {
	sortTasksByID(data.Tasks)
	payload, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("数据序列化失败: %w", err)
	}
	payload = append(payload, '\n')

	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("数据目录不可写: %w", err)
	}

	tempFile, err := os.CreateTemp(dir, ".todo-*.tmp")
	if err != nil {
		return fmt.Errorf("数据文件不可写: %w", err)
	}
	tempPath := tempFile.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tempPath)
		}
	}()

	if _, err := tempFile.Write(payload); err != nil {
		_ = tempFile.Close()
		return fmt.Errorf("数据文件不可写: %w", err)
	}
	if err := tempFile.Sync(); err != nil {
		_ = tempFile.Close()
		return fmt.Errorf("数据文件同步失败: %w", err)
	}
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("数据文件关闭失败: %w", err)
	}

	if err := os.Rename(tempPath, s.path); err != nil {
		return fmt.Errorf("数据文件原子更新失败: %w", err)
	}
	cleanup = false
	return nil
}

func (s *Store) Add(title string) (Task, error) {
	title, err := normalizeTitle(title)
	if err != nil {
		return Task{}, err
	}

	next := cloneData(s.data)
	now := s.now()
	task := Task{
		ID:        next.NextID,
		Title:     title,
		Status:    statusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}
	next.NextID++
	next.Tasks = append(next.Tasks, task)
	if err := s.saveData(next); err != nil {
		return Task{}, err
	}
	s.data = next
	return task, nil
}

func (s *Store) List(status string) ([]Task, error) {
	status, err := normalizeStatusFilter(status)
	if err != nil {
		return nil, err
	}

	tasks := make([]Task, 0, len(s.data.Tasks))
	for _, task := range s.data.Tasks {
		if status == "all" || task.Status == status {
			tasks = append(tasks, task)
		}
	}
	sort.SliceStable(tasks, func(i, j int) bool {
		if tasks[i].Status != tasks[j].Status {
			return tasks[i].Status == statusPending
		}
		return tasks[i].ID < tasks[j].ID
	})
	return tasks, nil
}

func (s *Store) Done(id int) error {
	next := cloneData(s.data)
	task := findTask(next.Tasks, id)
	if task == nil {
		return errTaskNotFound
	}
	if task.Status != statusDone {
		task.Status = statusDone
		task.UpdatedAt = s.now()
		if err := s.saveData(next); err != nil {
			return err
		}
		s.data = next
	}
	return nil
}

func (s *Store) Undo(id int) error {
	next := cloneData(s.data)
	task := findTask(next.Tasks, id)
	if task == nil {
		return errTaskNotFound
	}
	if task.Status != statusPending {
		task.Status = statusPending
		task.UpdatedAt = s.now()
		if err := s.saveData(next); err != nil {
			return err
		}
		s.data = next
	}
	return nil
}

func (s *Store) Edit(id int, title string) error {
	next := cloneData(s.data)
	task := findTask(next.Tasks, id)
	if task == nil {
		return errTaskNotFound
	}
	title, err := normalizeTitle(title)
	if err != nil {
		return err
	}
	task.Title = title
	task.UpdatedAt = s.now()
	if err := s.saveData(next); err != nil {
		return err
	}
	s.data = next
	return nil
}

func (s *Store) Delete(id int) error {
	next := cloneData(s.data)
	for i, task := range next.Tasks {
		if task.ID == id {
			next.Tasks = append(next.Tasks[:i], next.Tasks[i+1:]...)
			if err := s.saveData(next); err != nil {
				return err
			}
			s.data = next
			return nil
		}
	}
	return errTaskNotFound
}

func findTask(tasks []Task, id int) *Task {
	for i := range tasks {
		if tasks[i].ID == id {
			return &tasks[i]
		}
	}
	return nil
}

func emptyData() DataFile {
	return DataFile{
		Version: dataVersion,
		NextID:  1,
		Tasks:   []Task{},
	}
}

func validateData(data DataFile) error {
	if data.Version != dataVersion {
		return fmt.Errorf("version 必须为 %d", dataVersion)
	}
	if data.NextID < 1 {
		return errors.New("nextId 必须大于等于 1")
	}
	seen := map[int]bool{}
	for _, task := range data.Tasks {
		if task.ID < 1 {
			return errors.New("任务 id 必须大于等于 1")
		}
		if seen[task.ID] {
			return fmt.Errorf("任务 id 重复: %d", task.ID)
		}
		seen[task.ID] = true
		if task.ID >= data.NextID {
			return fmt.Errorf("nextId 必须大于已有任务 id")
		}
		if _, err := normalizeTitle(task.Title); err != nil {
			return fmt.Errorf("任务 %d 标题无效: %w", task.ID, err)
		}
		if task.Status != statusPending && task.Status != statusDone {
			return fmt.Errorf("任务 %d 状态无效", task.ID)
		}
		if task.CreatedAt.IsZero() || task.UpdatedAt.IsZero() {
			return fmt.Errorf("任务 %d 时间字段不能为空", task.ID)
		}
	}
	return nil
}

func normalizeTitle(title string) (string, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return "", errTitleEmpty
	}
	if utf8.RuneCountInString(title) > maxTitleRunes {
		return "", errTitleTooLong
	}
	for _, r := range title {
		if unicode.IsControl(r) {
			return "", errTitleControl
		}
	}
	return title, nil
}

func normalizeStatusFilter(status string) (string, error) {
	if status == "" {
		status = "all"
	}
	if status != "all" && status != statusPending && status != statusDone {
		return "", errInvalidStatus
	}
	return status, nil
}

func cloneData(data DataFile) DataFile {
	clone := data
	clone.Tasks = append([]Task(nil), data.Tasks...)
	return clone
}

func sortTasksByID(tasks []Task) {
	sort.SliceStable(tasks, func(i, j int) bool {
		return tasks[i].ID < tasks[j].ID
	})
}
