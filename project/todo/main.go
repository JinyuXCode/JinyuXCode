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

const dataFileName = "todo.json"

type appError struct {
	code     int
	message  string
	showHelp bool
}

func (e appError) Error() string {
	return e.message
}

func main() {
	code := run(os.Args[1:], os.Stdout, os.Stderr)
	os.Exit(code)
}

func run(args []string, stdout, stderr io.Writer) int {
	return runWithPath(args, filepath.Join(".", dataFileName), stdout, stderr)
}

func runWithPath(args []string, dataPath string, stdout, stderr io.Writer) int {
	if err := execute(args, dataPath, stdout); err != nil {
		var appErr appError
		if errors.As(err, &appErr) {
			fmt.Fprintln(stderr, appErr.message)
			if appErr.showHelp {
				fmt.Fprintln(stderr)
				printHelp(stderr)
			}
			return appErr.code
		}
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}

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

func loadStore(dataPath string) (*Store, error) {
	store := NewStore(dataPath)
	if err := store.Load(); err != nil {
		return nil, appError{code: 1, message: err.Error()}
	}
	return store, nil
}

func handleAdd(store *Store, title string, stdout io.Writer) error {
	task, err := store.Add(title)
	if err != nil {
		return businessError(err)
	}
	fmt.Fprintf(stdout, "已新增任务: %d\n", task.ID)
	return nil
}

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

func parseListArgs(args []string) (string, error) {
	flags := flag.NewFlagSet("list", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	status := flags.String("status", "all", "pending、done 或 all")
	if err := flags.Parse(args); err != nil {
		return "", appError{code: 2, message: "list 参数错误", showHelp: true}
	}
	if flags.NArg() > 0 {
		return "", appError{code: 2, message: "list 不接受多余参数", showHelp: true}
	}
	statusFilter, err := normalizeStatusFilter(*status)
	return statusFilter, businessError(err)
}

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

func parseTitleArgs(missingMessage string, args []string) (string, error) {
	if len(args) == 0 {
		return "", appError{code: 2, message: missingMessage, showHelp: true}
	}
	title, err := normalizeTitle(strings.Join(args, " "))
	return title, businessError(err)
}

func parseOnlyID(command string, args []string) (int, error) {
	if len(args) != 1 {
		return 0, appError{code: 2, message: command + " 需要且只接受一个 id", showHelp: true}
	}
	return parseID(args[0])
}

func parseID(raw string) (int, error) {
	id, err := strconv.Atoi(raw)
	if err != nil || id < 1 {
		return 0, appError{code: 2, message: "id 必须是正整数"}
	}
	return id, nil
}

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

const timeFormat = "2006-01-02T15:04:05Z07:00"

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
