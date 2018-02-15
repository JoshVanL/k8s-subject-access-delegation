package end_to_end

func initialStartup() (testBlock *TestBlock, err error) {
	programs := []*CommandArguments{
		&CommandArguments{
			program: "make",
			arguments: []string{
				"build_linux_sad",
			},
		},
		&CommandArguments{
			program: "minikube",
			arguments: []string{
				"start",
				"--extra-config=apiserver.Authorization.Mode=RBAC",
				"--memory=2048",
			},
		},
		&CommandArguments{
			program: "kubectl",
			arguments: []string{
				"get",
				"all",
			},
		},
	}

	return NewTestBlock("Initiating startup", programs)
}

func cleanUp() (testBlock *TestBlock, err error) {
	programs := []*CommandArguments{
		&CommandArguments{
			program: "minikube",
			arguments: []string{
				"delete",
			},
		},
	}

	return NewTestBlock("Cleaning up", programs)
}
