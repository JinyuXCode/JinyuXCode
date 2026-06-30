package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// 数据文件名，默认保存在当前工作目录下。
const dataFileName = "todo.json"

// appError 是程序自定义的错误类型。
// 它除了携带错误信息，还附带退出码（code）和是否需要打印帮助（showHelp），
// 上层据此决定如何向用户反馈、以什么退出码结束。
type appError struct {
	code     int
	message  string
	showHelp bool
}

// Error 实现 error 接口，返回给用户看的提示信息。
func (e appError) Error() string {
	return e.message
}

// main 是程序入口。真正的逻辑放在 run 中，main 只负责调用并退出。
// 这样拆分是为了在测试里可以直接调用 run，而不必启动整个进程。
func main() {
	code := run(os.Args[1:], os.Stdout, os.Stderr)
	os.Exit(code)
}

// run 用默认数据文件路径执行程序。
func run(args []string, stdout, stderr io.Writer) int {
	return runWithPath(args, filepath.Join(".", dataFileName), stdout, stderr)
}

// runWithPath 是程序的核心调度入口：执行命令，并根据返回的错误决定退出码。
// 通过注入 stdout/stderr，方便测试时捕获输出。
func runWithPath(args []string, dataPath string, stdout, stderr io.Writer) int {
	if err := execute(args, dataPath, stdout); err != nil {
		var appErr appError
		// 如果是我们自定义的 appError，按其携带的退出码退出，并按需打印帮助。
		if errors.As(err, &appErr) {
			fmt.Fprintln(stderr, appErr.message)
			if appErr.showHelp {
				fmt.Fprintln(stderr)
				printHelp(stderr)
			}
			return appErr.code
		}
		// 其它未预期错误统一以退出码 1 返回。
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}

// execute 负责解析并分发子命令（add/list/done/undo/edit/delete）。
// 每个分支套路一致：解析参数 → 加载数据 → 执行业务 → 返回错误。
func execute(args []string, dataPath string, stdout io.Writer) error {
	if len(args) == 0 {
		return appError{code: 2, message: "缺少命令", showHelp: true}
	}

	command := args[0]
	if command == "help" || command == "-h" || command == "--help" {
		printHelp(stdout)
		return nil
	}

	switch command {
	case "add":
		title, err := parseTitleArgs("缺少任务标题", args[1:])
		if err != nil {
			return err
		}
		store, err := loadStore(dataPath)
		if err != nil {
			return err
		}
		return handleAdd(store, title, stdout)
	case "list":
		status, err := parseListArgs(args[1:])
		if err != nil {
			return err
		}
		store, err := loadStore(dataPath)
		if err != nil {
			return err
		}
		return handleList(store, status, stdout)
	case "done":
		id, err := parseOnlyID("done", args[1:])
		if err != nil {
			return err
		}
		store, err := loadStore(dataPath)
		if err != nil {
			return err
		}
		return businessError(store.Done(id))
	case "undo":
		id, err := parseOnlyID("undo", args[1:])
		if err != nil {
			return err
		}
		store, err := loadStore(dataPath)
		if err != nil {
			return err
		}
		return businessError(store.Undo(id))
	case "edit":
		id, title, err := parseEditArgs(args[1:])
		if err != nil {
			return err
		}
		store, err := loadStore(dataPath)
		if err != nil {
			return err
		}
		return businessError(store.Edit(id, title))
	case "delete":
		id, err := parseOnlyID("delete", args[1:])
		if err != nil {
			return err
		}
		store, err := loadStore(dataPath)
		if err != nil {
			return err
		}
		return businessError(store.Delete(id))
	default:
		return appError{code: 2, message: "未知命令: " + command, showHelp: true}
	}
}

// loadStore 构造一个 Store 并从磁盘加载数据。
// 加载失败时包装成 appError（退出码 1，属于系统错误）。
func loadStore(dataPath string) (*Store, error) {
	store := NewStore(dataPath)
	if err := store.Load(); err != nil {
		return nil, appError{code: 1, message: err.Error()}
	}
	return store, nil
}

// handleAdd 处理 add 命令：新增任务并打印新任务的 ID。
func handleAdd(store *Store, title string, stdout io.Writer) error {
	task, err := store.Add(title)
	if err != nil {
		return businessError(err)
	}
	fmt.Fprintf(stdout, "已新增任务: %d\n", task.ID)
	return nil
}

// handleList 处理 list 命令：按状态过滤后以表格形式输出任务列表。
func handleList(store *Store, status string, stdout io.Writer) error {
	tasks, err := store.List(status)
	if err != nil {
		return businessError(err)
	}
	if len(tasks) == 0 {
		fmt.Fprintln(stdout, "暂无任务")
		return nil
	}

	fmt.Fprintln(stdout, "ID\tSTATUS\tTITLE\tCREATED_AT\tUPDATED_AT")
	for _, task := range tasks {
		fmt.Fprintf(stdout, "%d\t%s\t%s\t%s\t%s\n",
			task.ID,
			task.Status,
			task.Title,
			task.CreatedAt.Format(timeFormat),
			task.UpdatedAt.Format(timeFormat),
		)
	}
	return nil
}

// parseListArgs 解析 list 命令的参数，支持 --status 过滤。
// 使用标准库的 flag 包解析选项，默认展示全部任务。
func parseListArgs(args []string) (string, error) {
	flags := flag.NewFlagSet("list", flag.ContinueOnError)
	flags.SetOutput(io.Discard) // 屏蔽 flag 包自带的错误输出，由我们自己提示
	status := flags.String("status", "all", "pending、done 或 all")
	if err := flags.Parse(args); err != nil {
		return "", appError{code: 2, message: "list 参数错误", showHelp: true}
	}
	if flags.NArg() > 0 {
		// list 不接受位置参数，多余参数视为用法错误。
		return "", appError{code: 2, message: "list 不接受多余参数", showHelp: true}
	}
	statusFilter, err := normalizeStatusFilter(*status)
	return statusFilter, businessError(err)
}

// parseEditArgs 解析 edit 命令：需要 id 和新标题（标题可由多个参数拼接而成）。
func parseEditArgs(args []string) (int, string, error) {
	if len(args) < 2 {
		return 0, "", appError{code: 2, message: "edit 需要 id 和新标题", showHelp: true}
	}
	id, err := parseID(args[0])
	if err != nil {
		return 0, "", err
	}
	title, err := parseTitleArgs("edit 需要 id 和新标题", args[1:])
	return id, title, err
}

// parseTitleArgs 把剩余参数用空格拼接成标题，并做规范化校验。
func parseTitleArgs(missingMessage string, args []string) (string, error) {
	if len(args) == 0 {
		return "", appError{code: 2, message: missingMessage, showHelp: true}
	}
	title, err := normalizeTitle(strings.Join(args, " "))
	return title, businessError(err)
}

// parseOnlyID 用于只接受一个 id 的命令（done/undo/delete）。
func parseOnlyID(command string, args []string) (int, error) {
	if len(args) != 1 {
		return 0, appError{code: 2, message: command + " 需要且只接受一个 id", showHelp: true}
	}
	return parseID(args[0])
}

// parseID 把字符串解析为正整数 id，非法值返回用法错误。
func parseID(raw string) (int, error) {
	id, err := strconv.Atoi(raw)
	if err != nil || id < 1 {
		return 0, appError{code: 2, message: "id 必须是正整数"}
	}
	return id, nil
}

// businessError 把 Store 层的业务错误翻译成 appError。
// 已知业务错误（空标题、标题过长、任务不存在等）用退出码 2（用户用法错误），
// 其它未知错误用退出码 1（系统错误）。
func businessError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, errTitleEmpty) ||
		errors.Is(err, errTitleTooLong) ||
		errors.Is(err, errTitleControl) ||
		errors.Is(err, errTaskNotFound) ||
		errors.Is(err, errInvalidStatus) {
		return appError{code: 2, message: err.Error()}
	}
	return appError{code: 1, message: err.Error()}
}

// timeFormat 是任务时间的输出格式（RFC3339 风格）。
const timeFormat = "2006-01-02T15:04:05Z07:00"

// printHelp 输出用法说明。
func printHelp(w io.Writer) {
	fmt.Fprint(w, `Todo List

用法:
  todo add <title>
  todo list [--status pending|done|all]
  todo done <id>
  todo undo <id>
  todo edit <id> <new-title>
  todo delete <id>
  todo help

说明:
  title 去掉首尾空白后必须为 1~200 个字符。
  list 默认展示全部任务，并按 pending 在前、done 在后、同状态 id 升序排序。
  数据默认保存到当前工作目录的 todo.json。
`)
}
