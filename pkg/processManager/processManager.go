package processManager

import (
	"bufio"
	"io"
	"nps-auth/pkg/logger"
	"os/exec"
	"sync"
	"syscall"
)

type ProcessState string

const (
	Starting ProcessState = "Starting" // 启动中
	Running  ProcessState = "Running"  // 运行中
	Stopped  ProcessState = "Stopped"  // 停止
)

var (
	log                = logger.GetLogger("process-manager")
	maxRestartAttempts = 3 // 最大重启次数
)

type CheckStdout func(string) ProcessState
type ProcessManager struct {
	command      string
	args         []string
	cmd          *exec.Cmd
	mu           sync.Mutex
	status       ProcessState // 使用 status 代替原来的 isRunning
	restartCount int          // 当前重启次数

	checkStdout CheckStdout // 用于检查进程输出是否是启动成功的函数

}

// NewProcessManager 创建一个新的进程管理器实例
func NewProcessManager(command string, args []string, checkStdout CheckStdout) *ProcessManager {
	pm := ProcessManager{
		command:     command,
		args:        args,
		status:      Stopped, // 初始化状态为 Stopped
		checkStdout: checkStdout,
	}
	pm.KillByCommand()

	return &pm
}

// Start 启动可执行文件，并监听其输出
func (pm *ProcessManager) DoStart(isInit bool) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.status != Stopped {
		log.Warn().Str("status", string(pm.status)).Msg("Process doStart ignore")
		return nil
	}

	// 重置重启次数
	if isInit {
		pm.restartCount = 0
	}

	pm.cmd = exec.Command(pm.command, pm.args...)
	pm.cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// 创建管道以捕获标准输出和标准错误
	stdoutPipe, err := pm.cmd.StdoutPipe()
	if err != nil {
		log.Error().Err(err).Msg("failed to create stdout pipe")
		return err
	}

	stderrPipe, err := pm.cmd.StderrPipe()
	if err != nil {
		log.Error().Err(err).Msg("failed to create stderr pipe")
		return err
	}

	// 启动进程
	err = pm.cmd.Start()
	if err != nil {
		log.Error().Err(err).Msg("failed to start process")
		return err
	}

	// 启动goroutine读取stdout和stderr
	go pm.readPipeOutput(stdoutPipe, "stdout")
	go pm.readPipeOutput(stderrPipe, "stderr")

	// 启动进程监控
	go pm.monitorProcess()

	pm.status = Starting // 更新状态为 Starting
	log.Info().Msg("Process starting ")

	return nil
}

// handleRestart 处理进程重启逻辑
func (pm *ProcessManager) handleRestart() {
	pm.mu.Lock()
	pm.status = Stopped

	isRestartAble := pm.restartCount <= maxRestartAttempts
	if isRestartAble {
		pm.restartCount = pm.restartCount + 1
	}
	pm.mu.Unlock()

	if isRestartAble {
		log.Warn().Int("RestartCount", pm.restartCount).Msg("Restarting process...")
		pm.DoStart(false) // 重启进程
	} else {
		log.Error().Msg("Max restart attempts reached, not restarting the process")
	}
}

// readPipeOutput 读取管道输出
func (pm *ProcessManager) readPipeOutput(pipe io.ReadCloser, pipeName string) {
	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		text := scanner.Text()
		log.Info().Str(pipeName, text).Msg("Process output")

		status := pm.checkStdout(text)

		if status == Running {
			pm.mu.Lock()
			pm.status = Running
			pm.restartCount = 0
			pm.mu.Unlock()
		}
		if status == Stopped {
			pm.handleRestart() // 调用重启处理函数
		}

	}

	if err := scanner.Err(); err != nil {
		log.Error().Err(err).Msg("Process error")
		pm.handleRestart() // 调用重启处理函数
	}
}

// monitorProcess 监控进程状态，如果进程退出则自动重启
func (pm *ProcessManager) monitorProcess() {
	err := pm.cmd.Wait() // 等待进程结束

	log.Error().Err(err).Msg("Process exited unexpectedly")
	pm.handleRestart() // 调用重启处理函数
}

// Stop 停止可执行文件
func (pm *ProcessManager) DoStop() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.status == Stopped {
		log.Warn().Msg("process is already stopped")
		return nil
	}

	err := pm.cmd.Process.Kill()
	if err != nil {
		log.Error().Err(err).Msg("failed to stop process")
		return err
	}

	pm.status = Stopped // 更新状态为 Stopped
	pm.restartCount = maxRestartAttempts + 1
	log.Info().Int("Pid", pm.cmd.Process.Pid).Msg("process stopped")
	return nil
}

// IsRunning 检查进程是否在运行
func (pm *ProcessManager) GetStatus() ProcessState {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	return pm.status
}

// KillByCommand 通过路径命令停止进程
func (pm *ProcessManager) KillByCommand() error {
	cmd := exec.Command("sh", "-c", "ps aux | grep "+pm.command+" | grep -v grep | awk '{print $2}' | xargs kill -9")
	err := cmd.Run()
	if err != nil {
		log.Error().Str("command", cmd.String()).Err(err).Msg("failed to KillByCommand process")
		// return err
	}
	return nil
}
