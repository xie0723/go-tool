package cmdlib

import (
	"bufio"
	"context"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
)

//执行操作系统脚本或命令所依赖的对象
type ConcurrencyRunAdapter interface {
	GetDir() (string, bool)     //获取命令执行路径，bool为false时不需要设置
	DealStdOut(string) error    //标准输出
	DealStdErr(string) error    //标准错误输出
	GetTimeOut() time.Duration  //超时时间
	GetEnvs() map[string]string //需要预置的环境变量
	CloseLogFile() error        //关闭日志文件
	ErrResult() error           //获取执行结果是否错误
}

type InstructionParamsSeparator interface {
	Separator() string // 自定义命令行参数分隔符
}

type InstructionLogPrint interface {
	IsLogPrint() bool // 自定义是否输出日志
}

func ReplaceCmdNameByEnv(name string, envs []string) string {
	var _envsMap = make(map[string]string)
	for _, _env := range envs {
		_envSplit := strings.Split(_env, "=")
		if len(_envSplit) < 2 {
			continue
		}
		_envsMap[_envSplit[0]] = strings.Join(_envSplit[1:], "=")
	}
	replaceReg := beego.AppConfig.String("executor::EnvReplaceReg")
	if replaceReg == "" {
		replaceReg = "[^$]*([$][{]([^}]*)[}])[^$]*"
	}
	reg := regexp.MustCompile(replaceReg)
	submatch := reg.FindAllStringSubmatch(name, -1)
	for _, v := range submatch {
		if len(v) != 3 {
			continue
		}
		name = strings.Replace(name, v[1], _envsMap[v[2]], -1)
	}
	if len(submatch) == 0 {
		return name
	}
	return ReplaceCmdNameByEnv(name, envs)
}

//执行操作系统脚本或命令的方法
func ConcurrencyRun(name string, adapter ConcurrencyRunAdapter) (err error) {
	defer adapter.CloseLogFile()
	ctx, cancel := context.WithTimeout(context.Background(), adapter.GetTimeOut()) //超时时间以秒为单位
	defer cancel()
	var envTmp = os.Environ()
	//添加基础环境变量（执行用例的）
	for k, v := range adapter.GetEnvs() {
		envTmp = append(envTmp, k+"="+v)
	}
	name = ReplaceCmdNameByEnv(name, envTmp)
	var cmdName string
	var cmdArgs []string
	if adp, ok := adapter.(InstructionParamsSeparator); ok && adp.Separator() != "" {
		_cmdAndArgs := strings.Split(strings.TrimSpace(name), adp.Separator())
		cmdName = _cmdAndArgs[0]
		cmdArgs = _cmdAndArgs[1:]
	} else {
		_cmdAndArgs := strings.Fields(name)
		cmdName = _cmdAndArgs[0]
		cmdArgs = _cmdAndArgs[1:]
	}
	var isPrintLog = true
	if adp, ok := adapter.(InstructionLogPrint); ok {
		isPrintLog = adp.IsLogPrint()
	}

	_cmd := exec.CommandContext(ctx, cmdName, cmdArgs...)
	//执行目录(shell脚本才需要设置)
	if shellScriptDir, ok := adapter.GetDir(); ok {
		logs.Debug("Set cmd dir: %s", shellScriptDir)
		_cmd.Dir = shellScriptDir
	}
	//继承当前环境的环境变量
	_cmd.Env = envTmp
	//启动执行
	var stdout, stderr io.ReadCloser
	if stdout, err = _cmd.StdoutPipe(); err != nil { //标准输出
		logs.Error("Get StdoutPipe field,", err.Error())
		return err
	}
	if stderr, err = _cmd.StderrPipe(); err != nil { //标准错误输出
		logs.Error("Get StderrPipe field,", err.Error())
		return err
	}
	logs.Info("Begin to exec command:", name)
	if err = _cmd.Start(); err != nil {
		logs.Error("start cmd:%s failed, %s", name, err.Error())
		return err
	}
	errreader := bufio.NewReader(stderr) //获取错误输出
	go func() {
		for {
			line, _err := errreader.ReadString('\n')
			if _err != nil || io.EOF == _err {
				break
			}
			_line := strings.Replace(line, "\n", "", -1)
			_line = strings.TrimSpace(_line)
			if isPrintLog {
				logs.Error("Cmd Err Out:", _line)
			}
			if __err := adapter.DealStdErr(_line); __err != nil {
				logs.Error("DealStdErr failed,", __err.Error())
				err = __err
			}
		}
	}()
	stdreader := bufio.NewReader(stdout) //获取标准输出
	for {
		line, _err := stdreader.ReadString('\n')
		if _err != nil || io.EOF == _err {
			break
		}
		_line := strings.Replace(line, "\n", "", -1)
		_line = strings.TrimSpace(_line)
		if isPrintLog {
			logs.Info("Cmd Out:", _line)
		}
		if __err := adapter.DealStdOut(_line); __err != nil {
			logs.Error("DealStdOut failed,", __err.Error())
			err = __err
		}
	}
	if _err := _cmd.Wait(); _err != nil {
		logs.Error("Exec command: %s failed, %s", name, _err.Error())
		return _err
	}
	//判断adapter执行结果是否ok
	if adapter.ErrResult() != nil {
		logs.Error("adapter.ErrResult:%s", adapter.ErrResult())
		return adapter.ErrResult()
	}
	logs.Info("End to exec command:", name)
	return err
}

//同步执行操作系统命令的方法（获取命令返回值）
//RunShellCommandString 执行shell命令，返回字符串
func RunShellCommandString(cmdStr []string) (out string, err error) {
	cmd := exec.Command(cmdStr[0], cmdStr[1:]...)
	var rbytes []byte
	if rbytes, err = cmd.Output(); err == nil {
		out = string(rbytes)
		out = strings.TrimSpace(out)
		return out, nil
	}
	logs.Error("run shell command[%s] failed, %s", cmdStr, err.Error())
	return "", err
}

// 同步执行操作系统命令的方法（获取命令返回值）
// RunShellCommandString 执行shell命令，返回[]byte（包含stderr 和 stdout）
func RunShellCommandBytes(cmdStr []string) (out []byte, err error) {
	cmd := exec.Command(cmdStr[0], cmdStr[1:]...)
	if out, err = cmd.CombinedOutput(); err == nil {
		// out = bytes.TrimSpace(out)
		return out, nil
	}
	logs.Error("run shell command[%s] failed, %s", cmdStr, err.Error())
	return out, err
}
