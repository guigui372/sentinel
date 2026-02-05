package ssh

import (
	"bytes"
	"context"
	"fmt"
	"golang.org/x/crypto/ssh"
	"os"
	"time"
)

// RunCommand 连接服务器并执行单条命令
// 参数：IP, 端口, 用户名, 密码(可选), 密钥路径(可选), 要执行的命令
// 返回：命令的输出结果, 错误信息
func RunCommand(ctx context.Context, host string, port int, user, password, keyPath, cmd string) (string, error) {

	// === 1. 准备“身份证” (AuthMethod) ===
	var authMethods []ssh.AuthMethod

	// 优先尝试使用密钥 (Private Key)
	if keyPath != "" {
		// 读取密钥文件
		keyBytes, err := os.ReadFile(keyPath)
		if err == nil {
			// 解析私钥
			signer, err := ssh.ParsePrivateKey(keyBytes)
			if err == nil {
				authMethods = append(authMethods, ssh.PublicKeys(signer))
			}
		}
	}

	// 如果有密码，也加进去作为备选
	if password != "" {
		authMethods = append(authMethods, ssh.Password(password))
	}

	// === 2. 配置 SSH 客户端 (ClientConfig) ===
	config := &ssh.ClientConfig{
		User: user,
		Auth: authMethods,
		// HostKeyCallback: 这里的逻辑是验证服务器指纹。
		// 在内网运维中，为了方便通常设为 InsecureIgnoreHostKey (不安全但省事，不验证对方是谁)
		// 生产环境如果严格，需要像 known_hosts 那样验证
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second, // 5秒连不上就挂断
	}

	// === 3. 拨打电话 (Dial) ===
	// 拼接地址：192.168.1.100:22
	addr := fmt.Sprintf("%s:%d", host, port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return "", fmt.Errorf("连接失败: %v", err)
	}
	// 记得挂电话！(函数结束时执行)
	defer client.Close()

	// === 4. 创建会话 (Session) ===
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("创建会话失败: %v", err)
	}
	defer session.Close()
	var b bytes.Buffer
	session.Stdout = &b
	session.Stderr = &b
	if err := session.Start(cmd); err != nil {
		return "", fmt.Errorf("启动命令失败: %v", err)
	}

	// 创建一个通道，用来通知“命令执行完了”
	done := make(chan error, 1)
	go func() {
		done <- session.Wait() // 等待命令结束，把结果扔进管道
	}()

	// Select: 赛跑机制
	// 看是 done 管道先响（做完了），还是 ctx.Done 管道先响（超时了）
	select {
	case err := <-done:
		// 正常结束
		if err != nil {
			return b.String(), fmt.Errorf("执行出错: %v", err)
		}
		return b.String(), nil

	case <-ctx.Done():
		// 圣旨到了：超时撤退！
		// 必须手动杀掉 session，否则远程服务器上的命令可能还在跑
		// 发送 SIGKILL 信号（不一定所有 SSH 都支持，但这是标准做法）
		session.Signal(ssh.SIGKILL)
		session.Close()
		return "", fmt.Errorf("任务被取消或超时: %v", ctx.Err())
	}
}
