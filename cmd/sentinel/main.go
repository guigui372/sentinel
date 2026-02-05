package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"Go-Ops-Sentinel/internal/config"
	"Go-Ops-Sentinel/internal/monitor" // <--- 引入 monitor
	"Go-Ops-Sentinel/internal/notifier"
	"Go-Ops-Sentinel/internal/ssh"
)

// ScanResult 结构体保持不变
type ScanResult struct {
	Host    string
	Task    string
	Success bool
	Output  string
}

func main() {
	// 1. 加载配置
	conf, err := config.LoadConfig("configs/config.yaml")
	if err != nil {
		log.Fatalf("配置加载失败: %v", err)
	}

	// 2. 把具体的巡检逻辑封装成一个匿名函数(闭包)，或者直接传函数
	// 这里我们告诉 monitor：每隔 conf.Global.Interval 秒，就运行一次在这个 {} 里的代码
	monitor.Start(conf.Global.Interval, func() {
		// === 这里的代码就是以前 main 函数里的核心逻辑 ===
		runScanTask(conf)
	})
}

// runScanTask 是具体的干活逻辑 (被提炼出来了)
func runScanTask(conf *config.Config) {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conf.Global.Timeout)*time.Second)
	defer cancel()
	var wg sync.WaitGroup
	resultsChan := make(chan ScanResult, len(conf.Servers)*len(conf.Tasks))

	// --- 生产者 ---
	for _, server := range conf.Servers {
		wg.Add(1)
		go func(s config.Server) {
			defer wg.Done()
			for _, task := range conf.Tasks {
				// 注意：这里要处理一下超时，可以用 conf.Global.Timeout，暂时简略传参
				output, err := ssh.RunCommand(ctx, s.Host, s.Port, s.User, s.Password, s.KeyPath, task.Command)

				res := ScanResult{Host: s.Host, Task: task.Name}
				if err != nil {
					res.Success = false
					res.Output = fmt.Sprintf("%v", err)
				} else {
					res.Success = true
					res.Output = output
				}
				resultsChan <- res
			}
		}(server)
	}

	// --- Closer ---
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// --- 消费者 ---
	var failCount int
	for res := range resultsChan {
		if res.Success {
			// 成功日志简化输出
			cleanOut := strings.TrimSpace(res.Output)
			lines := strings.Split(cleanOut, "\n")
			info := ""
			if len(lines) > 0 {
				info = lines[0]
			}
			fmt.Printf("✅ [%s] %s: %s\n", res.Host, res.Task, info)
		} else {
			failCount++
			fmt.Printf("❌ [%s] %s: %s\n", res.Host, res.Task, res.Output)

			// 发送告警
			errMsg := fmt.Sprintf("服务器: %s\n任务: %s\n错误: %s", res.Host, res.Task, res.Output)
			notifier.SendDingTalk(conf.Global.Webhook, errMsg)
		}
	}

	fmt.Printf("--- 本轮巡检结束，耗时: %s, 异常数: %d ---\n", time.Since(start), failCount)
}
