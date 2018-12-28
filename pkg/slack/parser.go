package slack

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/infracloudio/kubeops/pkg/config"
	log "github.com/infracloudio/kubeops/pkg/logging"
)

var AllowedKubectlCommands = map[string]bool{
	"api-resources": true,
	"api-versions":  true,
	"cluster-info":  true,
	"describe":      true,
	"diff":          true,
	"explain":       true,
	"get":           true,
	"logs":          true,
	"set":           true,
	"top":           true,
	"version":       true,
	"auth":          true,
}

var AllowedNotifierCommands = map[string]bool{
	"notifier": true,
	"help":     true,
	"ping":     true,
}

func ParseAndRunCommand(msg string) string {
	args := strings.Split(msg, " ")
	if AllowedKubectlCommands[args[0]] {
		return runKubectlCommand(args)
	}
	if AllowedNotifierCommands[args[0]] {
		return runNotifierCommand(args)
	}
	return "Command not supported. Please run '@kubeops help' to see supported commands"
}

func runKubectlCommand(args []string) string {
	// Use 'default' as a default namespace
	args = append([]string{"-n", "default"}, args...)

	// Remove unnecessary flags
	finalArgs := []string{}
	for _, a := range args {
		if a == "-f" || strings.HasPrefix(a, "--follow") {
			continue
		}
		if a == "-w" || strings.HasPrefix(a, "--watch") {
			continue
		}
		finalArgs = append(finalArgs, a)
	}

	cmd := exec.Command("/usr/local/bin/kubectl", finalArgs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Logger.Error("Error in executing kubectl command: ", err)
		return string(out) + err.Error()
	}
	return string(out)
}

// TODO: Have a seperate cli which runs bot commands
func runNotifierCommand(args []string) string {
	switch len(args) {
	case 1:
		if strings.ToLower(args[0]) == "help" {
			return printHelp()
		}
		if strings.ToLower(args[0]) == "ping" {
			return "pong"
		}
	case 2:
		if args[0] != "notifier" {
			return printDefaultMsg()
		}
		if args[1] == "start" {
			config.Notify = true
			return "Notifier started!"
		}
		if args[1] == "stop" {
			config.Notify = false
			return "Notifier stopped!"
		}
		if args[1] == "status" {
			if config.Notify == false {
				return "stopped"
			}
			return "running"
		}
		if args[1] == "showconfig" {
			out, err := showControllerConfig()
			if err != nil {
				log.Logger.Error("Error in executing showconfig command: ", err)
				return "Error in getting configuration!"
			}
			return out
		}
	}
	return printDefaultMsg()
}

func printHelp() string {
	allowedKubectl := ""
	for k, _ := range AllowedKubectlCommands {
		allowedKubectl = allowedKubectl + k + ", "
	}
	helpMsg := "kubeops executes kubectl commands on k8s cluster and returns output.\n" +
		"Usages:\n" +
		"    @kubeops <kubectl command without `kubectl` prefix>\n" +
		"e.g:\n" +
		"    @kubeops get pods\n" +
		"    @kubeops logs podname -n namespace\n" +
		"Allowed kubectl commands:\n" +
		"    " + allowedKubectl + "\n\n" +
		"Commands to manage notifier:\n" +
		"notifier stop          Stop sending k8s event notifications to slack (started by default)\n" +
		"notifier start         Start sending k8s event notifications to slack\n" +
		"notifier status        Show running status of event notifier\n" +
		"notifier showconfig    Show kubeops configuration for event notifier\n\n" +
		"Other Commands:\n" +
		"help                   Show help\n" +
		"ping                   Check connection health\n"
	return helpMsg

}

func printDefaultMsg() string {
	return "Command not supported. Please run '@kubeops help' to see supported commands"
}

func showControllerConfig() (string, error) {
	configPath := os.Getenv("KUBEOPS_CONFIG_PATH")
	configFile := filepath.Join(configPath, config.ConfigFileName)
	file, err := os.Open(configFile)
	defer file.Close()
	if err != nil {
		return "", err
	}

	b, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}

	return string(b), nil
}
