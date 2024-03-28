package main

func InitFctl() {
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

	// write default config into fctl home
	defaultConf, err := defaultFctlConfig()
	if err != nil {
		logs.Printf("get default config failed: %v", err)
		return
	}
	if err := saveFctlConfig(defaultConf); err != nil {
		logs.Printf("save fctl config failed: %v", err)
		return
	}
}
