package main

func initFctl() {
	err := Run("rm", "-rf", FctlHome)
	if err != nil {
		logs.Printf("remove old FctlHome directory failed: %v", err)
		return
	}

	err = Run("mkdir", "-p", "-m", "700", FctlHome)
	if err != nil {
		logs.Printf("create new FctlHome directory failed: %v", err)
		return
	}

	// download monitor
	err = FctlRun("wget", "--timeout=10", `http://github.com/TTraveller7/invokerlib-monitor/archive/main.tar.gz`)
	if err != nil {
		logs.Printf("download monitor failed: %v", err)
		return
	}
	defer FctlRun("rm", "main.tar.gz")

	err = FctlRun("tar", "-xvf", "main.tar.gz")
	if err != nil {
		logs.Printf("extract monitor failed: %v", err)
		return
	}
}
