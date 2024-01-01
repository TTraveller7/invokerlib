package fctl

func initFctl() {
	err := Run("rm", "-rf", FctlHome)
	if err != nil {
		logs.Printf("remove old FctlHome directory failed: %v", err)
		return
	}

	err = Run("mkdir", "-p", "-m", "777", FctlHome)
	if err != nil {
		logs.Printf("create new FctlHome directory failed: %v", err)
		return
	}

	err = Run("cd", FctlHome)
	if err != nil {
		logs.Printf("change directory to FctlHome failed: %v", err)
		return
	}

	// download monitor
	err = Run("wget", `http://github.com/TTraveller7/invokerlib-monitor/archive/main.tar.gz`)
	if err != nil {
		logs.Printf("download monitor failed: %v", err)
		return
	}
	defer Run("rm", "main.tar.gz")

	err = Run("tar", "-xvf", "main.tar.gz")
	if err != nil {
		logs.Printf("extract monitor failed: %v", err)
		return
	}
}
