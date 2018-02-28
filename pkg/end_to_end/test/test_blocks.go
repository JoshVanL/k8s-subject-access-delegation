package test

import (
	"github.com/hashicorp/go-multierror"
)

func (t *TestingSuite) allTestBlocks() (blocks []*TestBlock, err error) {
	var result *multierror.Error

	block, err := initialStartup()
	if err != nil {
		result = multierror.Append(result, err)
	}
	blocks = append(blocks, block)

	block, err = test1()
	if err != nil {
		result = multierror.Append(result, err)
	}
	blocks = append(blocks, block)

	block, err = cleanUp()
	if err != nil {
		result = multierror.Append(result, err)
	}
	blocks = append(blocks, block)

	return blocks, result.ErrorOrNil()
}

func initialStartup() (testBlock *TestBlock, err error) {
	cmds := []*Command{
		&Command{
			program: "minikube",
			arguments: []string{
				"start",
				"--extra-config=apiserver.Authorization.Mode=RBAC",
				"--memory=1024",
			},
			background: false,
		},
		&Command{
			program: "kubectl",
			arguments: []string{
				"get",
				"all",
			},
			background: false,
		},
		&Command{
			program:    "./k8s_subject_access_delegation_linux_amd64",
			arguments:  []string{},
			background: true,
		},
		&Command{
			program: "kubectl",
			arguments: []string{
				"create",
				"-f",
				"docs/pod-role-service-account.yaml",
			},
			delay:      4,
			background: false,
		},
	}

	return NewTestBlock("Initiating startup", cmds)
}

func test1() (testBlock *TestBlock, err error) {
	cmds := []*Command{
		&Command{
			program: "kubectl",
			arguments: []string{
				"create",
				"-f",
				"docs/e2e_1.yaml",
			},
			background: false,
		},
		&Command{
			program: "kubectl",
			arguments: []string{
				"create",
				"-f",
				"docs/nginx_pod.yaml",
			},
			background: false,
			delay:      1,
		},
		&Command{
			program: "kubectl",
			arguments: []string{
				"get",
				"rolebindings",
			},
			background: false,
			delay:      1,
			conditions: []Condition{
				&SplitStringCondition{
					line:  1,
					split: 0,
					match: "test-pod-add-default-pod-reader",
				},
			},
		},
		&Command{
			program: "kubectl",
			arguments: []string{
				"get",
				"rolebindings",
			},
			background: false,
			delay:      5,
			conditions: []Condition{
				&StringCondition{
					line:  0,
					match: "No resources found.",
				},
			},
		},
	}

	return NewTestBlock("test1", cmds)
}

func cleanUp() (testBlock *TestBlock, err error) {
	cmds := []*Command{
		&Command{
			program: "minikube",
			arguments: []string{
				"delete",
			},
		},
	}

	return NewTestBlock("Cleaning up", cmds)
}
