package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/TTraveller7/invokerlib"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

func Create() {
	pathPtr := pflag.StringP("config", "c", "", "Yaml config path.")
	keepAliveOnFailurePtr := pflag.BoolP("keepAliveOnFailure", "k", false,
		"If set to true, resources are kept alive and not removed when any creation step fails. Defualt false.")

	pflag.Parse()
	if pathPtr == nil || len(*pathPtr) == 0 {
		logs.Printf("config path is not provided. Use -c <config path> to provide config path. ")
		return
	}
	keepAliveOnFailure := true
	if keepAliveOnFailurePtr != nil {
		keepAliveOnFailure = *keepAliveOnFailurePtr
	}

	// TODO: check fission status

	// parse config yaml
	content, err := os.ReadFile(*pathPtr)
	if err != nil {
		logs.Printf("read config file failed: %v", err)
		return
	}

	invokerConfig := &invokerlib.RootConfig{}
	if err := yaml.Unmarshal(content, invokerConfig); err != nil {
		logs.Printf("unmarshal config file failed: %v", err)
		return
	}

	conf.RootConfigPath = *pathPtr
	if err := saveFctlConfig(conf); err != nil {
		logs.Printf("save fctl config failed: %v", err)
		return
	}

	// validate config
	if err := invokerConfig.Validate(); err != nil {
		logs.Printf("validate config failed: %v", err)
		return
	}

	fissionStartSuccess := false

	// create fission env
	err = Run("fission", "env", "create",
		"--name", FissionEnv,
		"--image", "ttraveller7/go-env-1.19",
		"--builder", "ttraveller7/go-builder-1.19",
		"--poolsize", "1",
		"--version", "3")
	if err != nil {
		logs.Printf("create fission env failed: %v", err)
		return
	}
	defer func() {
		if !keepAliveOnFailure && !fissionStartSuccess {
			Run("fission", "env", "delete",
				"--name", FissionEnv)
		}
	}()

	// create monitor function
	err = Run("fission", "fn", "create",
		"--name", "monitor",
		"--env", FissionEnv,
		"--entrypoint", "Handler",
		"--executortype", "newdeploy",
		"--minscale", "1",
		"--maxscale", "1",
		"--src", ConcatPath(conf.MonitorDirectoryPath, "go.mod"),
		"--src", ConcatPath(conf.MonitorDirectoryPath, "go.sum"),
		"--src", ConcatPath(conf.MonitorDirectoryPath, "handler.go"))
	if err != nil {
		logs.Printf("create monitor function failed: %v", err)
		return
	}
	defer func() {
		if !keepAliveOnFailure && !fissionStartSuccess {
			Run("fission", "fn", "delete",
				"--name", "monitor")
		}
	}()

	// create monitor http endpoint
	err = Run("fission", "httptrigger", "create",
		"--name", "monitor",
		"--url", "/monitor",
		"--method", "POST",
		"--function", "monitor")
	if err != nil {
		logs.Printf("create monitor httptrigger failed: %v", err)
		return
	}
	defer func() {
		if !keepAliveOnFailure && !fissionStartSuccess {
			Run("fission", "httptrigger", "delete",
				"--name", "monitor")
		}
	}()

	time.Sleep(1 * time.Second)

	cli := NewMonitorClient()

	// load monitor config
	logs.Printf("sending command loadRootConfig to monitor")
	resp, err := cli.LoadRootConfig(invokerConfig)
	if err != nil {
		logs.Printf("loadRootConfig failed: %v", err)
		return
	} else if resp.Code != invokerlib.ResponseCodes.Success {
		logs.Printf("loadRootConfig failed with resp: %+v", resp)
		return
	}
	logs.Printf("loadRootConfig finished with resp: %+v", resp)

	// create topics from root config
	logs.Printf("sending command createTopics to monitor")
	resp, err = cli.CreateTopics()
	if err != nil {
		logs.Printf("createTopics failed: %v", err)
		return
	} else if resp.Code != invokerlib.ResponseCodes.Success {
		logs.Printf("createTopics failed with resp: %+v", resp)
		return
	}
	logs.Printf("createTopics finished with resp: %+v", resp)

	// create processors
	logs.Printf("create processors")
	for _, processorConf := range invokerConfig.ProcessorConfigs {
		name := processorConf.Name
		cmd := []string{"fission", "fn", "create",
			"--name", name,
			"--env", FissionEnv,
			"--entrypoint", processorConf.EntryPoint,
			"--executortype", "newdeploy",
			"--minscale", "1",
			"--maxscale", "1",
		}
		for _, file := range processorConf.Files {
			cmd = append(cmd, "--src")
			cmd = append(cmd, ConcatPath(processorConf.ParentDirectory, file))
		}
		err = Run(cmd...)
		if err != nil {
			logs.Printf("create processor %s failed: %v", name, err)
			return
		}
		defer func() {
			if !keepAliveOnFailure && !fissionStartSuccess {
				Run("fission", "fn", "delete",
					"--name", name)
			}
		}()

		// create processor http endpoint
		err = Run("fission", "httptrigger", "create",
			"--name", name,
			"--url", "/"+name,
			"--method", "POST",
			"--function", name)
		if err != nil {
			logs.Printf("create monitor httptrigger failed: %v", err)
			return
		}
		defer func() {
			if !keepAliveOnFailure && !fissionStartSuccess {
				Run("fission", "httptrigger", "delete",
					"--name", name)
			}
		}()

		time.Sleep(1 * time.Second)

		pc := NewProcessorClient(name)
		resp, err := pc.Ping()
		if err != nil {
			logs.Printf("ping processor client %s failed: %v", name, err)
			return
		} else if resp.Code != invokerlib.ResponseCodes.Success {
			logs.Printf("ping processor client %s failed with resp: %+v", name, resp)
			return
		}
	}

	// gather k8s services of processors, pass the IPs to monitor
	output, err := Exec("kubectl", "get", "svc",
		"-o", "name")
	if err != nil {
		logs.Printf("get k8s service failed: %v", err)
		return
	}
	output = strings.Trim(output, "\n")
	splittedOutput := strings.Split(output, "\n")
	serviceNames := make([]string, 0)
	for _, svcName := range splittedOutput {
		if strings.Contains(svcName, "newdeploy-") {
			n := strings.TrimPrefix(svcName, "service/")
			n = fmt.Sprintf("http://%s", n)
			serviceNames = append(serviceNames, n)
		}
	}

	// pass endpoints to monitor
	p := &invokerlib.LoadProcessorEndpointsParams{
		Endpoints: serviceNames,
	}
	logs.Printf("sending command loadProcessorEndpoints to monitor")
	resp, err = cli.LoadProcessorEndpoints(p)
	if err != nil {
		logs.Printf("loadProcessorEndpoints failed: %v", err)
		return
	} else if resp.Code != invokerlib.ResponseCodes.Success {
		logs.Printf("loadProcessorEndpoints failed with resp: %+v", resp)
		return
	}
	logs.Printf("loadProcessorEndpoints finished with resp: %+v", resp)

	// initialize processors
	logs.Printf("sending command initializeProcessors to monitor")
	resp, err = cli.InitializeProcessors()
	if err != nil {
		logs.Printf("initializeProcessors failed: %v", err)
		return
	} else if resp.Code != invokerlib.ResponseCodes.Success {
		logs.Printf("initializeProcessors failed with resp: %+v", resp)
		return
	}
	logs.Printf("initializeProcessors finished with resp: %+v", resp)

	fissionStartSuccess = true
}
