package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/RichXan/xcommon/xoauth"
)

const (
	defaultKeysDir = "./keys"
	// ANSI 颜色代码
	colorGreen = "\033[0;32m"
	colorRed   = "\033[0;31m"
	colorReset = "\033[0m"
)

func main() {
	// 获取密钥目录
	keysDir := defaultKeysDir
	if len(os.Args) > 1 {
		keysDir = os.Args[1]
	}

	fmt.Printf("%s开始生成密钥对...%s\n", colorGreen, colorReset)

	// 检查并创建目录
	if err := os.MkdirAll(keysDir, 0700); err != nil {
		log.Fatalf("%s创建目录失败: %s: %v%s\n", colorRed, keysDir, err, colorReset)
	}

	// 生成密钥对
	claims := xoauth.NewClaims(nil)
	if err := claims.GenerateKeyPair(keysDir); err != nil {
		log.Fatalf("%s生成密钥对失败: %v%s\n", colorRed, err, colorReset)
	}

	// 验证文件是否生成
	privateKeyPath := filepath.Join(keysDir, "private.pem")
	publicKeyPath := filepath.Join(keysDir, "public.pem")

	if _, err := os.Stat(privateKeyPath); err != nil {
		log.Fatalf("%s私钥文件未生成: %v%s\n", colorRed, err, colorReset)
	}
	if _, err := os.Stat(publicKeyPath); err != nil {
		log.Fatalf("%s公钥文件未生成: %v%s\n", colorRed, err, colorReset)
	}

	// 显示成功信息
	fmt.Printf("%s密钥对生成成功！%s\n", colorGreen, colorReset)
	fmt.Printf("私钥文件: %s%s%s\n", colorGreen, privateKeyPath, colorReset)
	fmt.Printf("公钥文件: %s%s%s\n", colorGreen, publicKeyPath, colorReset)

	// 显示文件权限
	if err := printFilePermissions(keysDir); err != nil {
		log.Printf("%s无法显示文件权限: %v%s\n", colorRed, err, colorReset)
	}
}

func printFilePermissions(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		// 模拟 ls -l 的输出格式
		fmt.Printf("%s %8d %s %s\n",
			info.Mode(),
			info.Size(),
			info.ModTime().Format("Jan 02 15:04"),
			info.Name())
	}
	return nil
}
