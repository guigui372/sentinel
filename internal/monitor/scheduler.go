package monitor

import (
	"fmt"
	"time"
)

// Start 启动调度器
// intervalSeconds: 间隔多少秒
// job: 一个函数，代表要具体干的活（在这个项目里就是巡检逻辑）
func Start(intervalSeconds int, job func()) {
	if intervalSeconds <= 0 {
		intervalSeconds = 60 // 默认保底 60秒
	}

	// 1. 创建打点器 (Ticker)
	// 就像一个节拍器，每隔 X 秒就会“叮”一声
	duration := time.Duration(intervalSeconds) * time.Second
	ticker := time.NewTicker(duration)

	fmt.Printf("🚀 监控服务已启动，巡检间隔: %d 秒\n", intervalSeconds)

	// 2. 立刻先跑一次 (不然启动后要干等60秒才会跑第一次)
	job()

	// 3. 死循环监听节拍器
	for range ticker.C {
		// 只要 ticker.C 管道里吐出一个时间点，就说明时间到了
		fmt.Println("\n⏰ 触发定时巡检...")
		job()
	}
}
