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

// 数据文件相关常量。
const (
	dataVersion   = 1         // 数据文件版本号，用于将来格式升级时做兼容判断
	statusPending = "pending" // 任务未完成状态
	statusDone    = "done"    // 任务已完成状态
	maxTitleRunes = 200       // 标题最大字符数（按 rune 计，兼容中文等非 ASCII）
)

// 业务层预定义的错误。用 errors.Is 判断类型，便于统一处理。
var (
	errTitleEmpty    = errors.New("标题不能为空")
	errTitleTooLong  = fmt.Errorf("标题不能超过 %d 个字符", maxTitleRunes)
	errTitleControl  = errors.New("标题不能包含控制字符")
	errTaskNotFound  = errors.New("任务不存在")
	errInvalidStatus = errors.New("status 仅支持 pending、done 或 all")
)

// Task 表示一条待办任务。
type Task struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// DataFile 是 todo.json 在内存中的结构，也是序列化到磁盘的格式。
type DataFile struct {
	Version int    `json:"version"` // 数据格式版本
	NextID  int    `json:"nextId"`  // 下一个可分配的任务 id
	Tasks   []Task `json:"tasks"`   // 全部任务列表
}

// Store 负责管理数据的加载与持久化。
// now 字段注入时间函数，方便测试时替换成固定时间。
type Store struct {
	path string
	now  func() time.Time
	data DataFile
}

// NewStore 创建一个指向指定数据文件的 Store。
func NewStore(path string) *Store {
	return &Store{
		path: path,
		now:  time.Now,
	}
}

// Load 从磁盘加载数据。
// 若文件不存在，则初始化一份空数据并落盘；若存在则解析、校验、排序后载入内存。
func (s *Store) Load() error {
	content, err := os.ReadFile(s.path)
	if err != nil {
		// 文件不存在是正常情况：创建一份空的初始数据并写盘。
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
	// 解析 JSON；解析失败通常意味着文件被损坏。
	if err := json.Unmarshal(content, &data); err != nil {
		return fmt.Errorf("数据文件解析失败，请备份或重建 %s: %w", filepath.Base(s.path), err)
	}

	// 校验数据内容是否合法（版本、id 唯一性、字段非空等）。
	if err := validateData(data); err != nil {
		return fmt.Errorf("数据文件内容无效，请备份或重建 %s: %w", filepath.Base(s.path), err)
	}

	// 按 id 排序，保证内存中数据顺序稳定。
	sortTasksByID(data.Tasks)
	s.data = data
	return nil
}

// Save 把当前内存数据写回磁盘。
func (s *Store) Save() error {
	data := cloneData(s.data)
	return s.saveData(data)
}

// saveData 是落盘的核心实现，采用"原子写入"策略：
// 先写临时文件并落盘，再用 os.Rename 原子替换原文件。
// 这样即使写入过程中崩溃，原文件也不会被破坏成半截。
func (s *Store) saveData(data DataFile) error {
	// 写入前先排序，保证磁盘上的任务顺序稳定。
	sortTasksByID(data.Tasks)
	payload, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("数据序列化失败: %w", err)
	}
	payload = append(payload, '\n') // 末尾加换行，便于命令行工具查看

	// 确保数据目录存在。
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("数据目录不可写: %w", err)
	}

	// 在同目录下创建临时文件，写完后用 Rename 替换原文件。
	// 同目录是为了保证 Rename 是原子操作（跨文件系统会退化为复制）。
	tempFile, err := os.CreateTemp(dir, ".todo-*.tmp")
	if err != nil {
		return fmt.Errorf("数据文件不可写: %w", err)
	}
	tempPath := tempFile.Name()
	// cleanup 标志用于 defer：正常完成时置为 false 跳过清理，出错时删除残留临时文件。
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tempPath)
		}
	}()

	// 写入数据。
	if _, err := tempFile.Write(payload); err != nil {
		_ = tempFile.Close()
		return fmt.Errorf("数据文件不可写: %w", err)
	}
	// Sync 强制把内核缓冲区刷到磁盘，避免断电丢数据。
	if err := tempFile.Sync(); err != nil {
		_ = tempFile.Close()
		return fmt.Errorf("数据文件同步失败: %w", err)
	}
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("数据文件关闭失败: %w", err)
	}

	// 原子替换：rename 成功后原文件名即指向新内容。
	if err := os.Rename(tempPath, s.path); err != nil {
		return fmt.Errorf("数据文件原子更新失败: %w", err)
	}
	cleanup = false // 临时文件已被 Rename 接管，不再需要清理
	return nil
}

// Add 新增一条任务，并返回新建的任务。
// 采用"克隆→修改克隆→落盘→提交"模式：先在副本上修改，落盘成功后才更新内存，
// 这样即使写盘失败，内存状态也保持一致。
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
	s.data = next // 落盘成功后才提交内存状态
	return task, nil
}

// List 按状态过滤任务，并按"pending 在前、done 在后、同状态 id 升序"排序后返回。
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
			return tasks[i].Status == statusPending // pending 排在 done 前面
		}
		return tasks[i].ID < tasks[j].ID
	})
	return tasks, nil
}

// Done 把指定任务标记为已完成。已经是 done 时直接返回，不更新时间。
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

// Undo 把指定任务恢复为未完成。
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

// Edit 修改指定任务的标题。
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

// Delete 删除指定任务。
func (s *Store) Delete(id int) error {
	next := cloneData(s.data)
	for i, task := range next.Tasks {
		if task.ID == id {
			// 用切片拼接删掉第 i 个元素。
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

// findTask 在任务列表中按 id 查找，返回指向该任务的指针（找不到返回 nil）。
// 返回指针是为了让调用方能直接修改找到的任务字段。
func findTask(tasks []Task, id int) *Task {
	for i := range tasks {
		if tasks[i].ID == id {
			return &tasks[i]
		}
	}
	return nil
}

// emptyData 返回一份合法的空初始数据。
func emptyData() DataFile {
	return DataFile{
		Version: dataVersion,
		NextID:  1,
		Tasks:   []Task{},
	}
}

// validateData 校验从磁盘读入的数据是否合法。
// 主要防止手工编辑或文件损坏导致程序行为异常。
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
		// 已有任务的 id 必须小于 nextId，否则会分配出重复 id。
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

// normalizeTitle 规范化并校验标题：去首尾空白，检查非空、长度、控制字符。
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

// normalizeStatusFilter 规范化状态过滤条件，空字符串视为 all。
func normalizeStatusFilter(status string) (string, error) {
	if status == "" {
		status = "all"
	}
	if status != "all" && status != statusPending && status != statusDone {
		return "", errInvalidStatus
	}
	return status, nil
}

// cloneData 深拷贝一份数据（重点是复制 Tasks 切片底层数组）。
// 这样在副本上修改不会影响内存中的原始数据，实现"写时复制"。
func cloneData(data DataFile) DataFile {
	clone := data
	clone.Tasks = append([]Task(nil), data.Tasks...)
	return clone
}

// sortTasksByID 按 id 升序稳定排序。
func sortTasksByID(tasks []Task) {
	sort.SliceStable(tasks, func(i, j int) bool {
		return tasks[i].ID < tasks[j].ID
	})
}
