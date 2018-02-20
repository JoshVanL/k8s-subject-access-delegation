package end_to_end

func initialStartup() (testBlock *TestBlock, err error) {
	programs := []*CommandArguments{
		//&CommandArguments{
		//	program: "make",
		//	arguments: []string{
		//		"build_linux_sad",
		//	},
		//},
		&CommandArguments{
			program: "minikube",
			arguments: []string{
				"start",
				"--extra-config=apiserver.Authorization.Mode=RBAC",
				"--memory=2048",
			},
			background: false,
		},
		&CommandArguments{
			program: "kubectl",
			arguments: []string{
				"get",
				"all",
			},
			background: false,
		},
		&CommandArguments{
			program:    "./k8s_subject_access_delegation_linux_amd64",
			arguments:  []string{},
			background: true,
		},
		&CommandArguments{
			program: "kubectl",
			arguments: []string{
				"create",
				"-f",
				"docs/pod-role-service-account.yaml",
			},
			background: false,
		},
		&CommandArguments{
			program: "kubectl",
			arguments: []string{
				"create",
				"-f",
				"docs/pod-role-service-account.yaml",
			},
			background: false,
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
