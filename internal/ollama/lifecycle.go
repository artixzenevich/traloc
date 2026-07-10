package ollama

import (
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/charmbracelet/log"
)

func IsRunning() bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://127.0.0.1:11434/api/tags")
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == 200
}

func waitForServer(timeout int) bool {
	for i := 0; i < timeout; i++ {
		if IsRunning() {
			return true
		}
		time.Sleep(1 * time.Second)
	}
	return false
}

func StartIfNeeded() (*exec.Cmd, error) {
	if IsRunning() {
		log.Info("Ollama уже запущен — используем существующий процесс")
		return nil, nil
	}

	log.Info("Запускаем Ollama...")
	cmd := exec.Command("ollama", "serve")
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("не удалось запустить ollama: %w", err)
	}

	if !waitForServer(30) {
		cmd.Process.Kill()
		return nil, fmt.Errorf("ollama не поднялся за 30 секунд")
	}

	log.Info("Ollama запущен")
	return cmd, nil
}

func Stop(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	log.Info("Останавливаем Ollama...")

	if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
		cmd.Process.Kill()
		return
	}

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case <-time.After(5 * time.Second):
		log.Warn("Ollama не завершился — убиваем")
		cmd.Process.Kill()
	case err := <-done:
		if err != nil {
			log.Infof("Ollama завершился: %v", err)
		} else {
			log.Info("Ollama остановлен")
		}
	}
}

func UnloadModel(model string) {
	log.Info("Выгружаем модель из VRAM...")
	body := fmt.Sprintf(`{"model":"%s","keep_alive":0}`, model)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(
		"http://127.0.0.1:11434/api/generate",
		"application/json",
		strings.NewReader(body),
	)
	if err != nil {
		log.Warnf("Не удалось выгрузить модель: %v", err)
		return
	}
	resp.Body.Close()
	log.Info("Модель выгружена")
}
