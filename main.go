package main

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

//go:embed index.html
var htmlContent []byte

type handlerWithPort struct {
	listenPort string
	htmlData   []byte
}

func (h *handlerWithPort) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("【访问日志】客户端IP:%s 访问本机端口:%s\n", r.RemoteAddr, h.listenPort)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(h.htmlData)
}

func startListen(port string, htmlData []byte) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/", &handlerWithPort{
		listenPort: port,
		htmlData:   htmlData,
	})

	server := &http.Server{
		Addr:         "0.0.0.0:" + port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		fmt.Printf("[监听成功] 0.0.0.0:%s\n", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("[监听失败] %s 端口: %v\n", port, err)
		}
	}()

	return server
}

func main() {
	fmt.Printf("HTML 大小: %d 字节\n", len(htmlContent))

	ports := []string{"80", "443", "8080"}
	var servers []*http.Server
	var wg sync.WaitGroup

	for _, p := range ports {
		wg.Add(1)
		srv := startListen(p, htmlContent)
		servers = append(servers, srv)
		go func(s *http.Server) {
			defer wg.Done()
			// 等待关闭信号
		}(srv)
	}

	// 监听退出信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\n[收到退出信号] 正在关闭服务...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, srv := range servers {
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("关闭失败: %v", err)
		}
	}

	fmt.Println("[所有服务已关闭]")
}
