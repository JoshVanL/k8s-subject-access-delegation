#!/bin/bash

RED='\e[1;31m'
BLU='\e[0;36m'
WIT='\e[0;0m'

uname=$(uname)
OS=${uname,,}

SADProcess="k8s_subject_access_delegation_${OS}_amd64"
CmdSleep="sleep 10"

CmdDelete="minikube delete"
CmdStart="minikube start --extra-config=apiserver.Authorization.Mode=RBAC --memory 2048"
CmdSAD="./${SADProcess} &"
CmdK8sGetAll="kubectl get all"

CmdCreateTestResources="kubectl create -f docs/testing_roles_service_accounts.yaml"
CmdTest1="kubectl create --filename docs/testing_sad_1.yaml"


func_print() {
    printf "\n> ${BLU}$1${WIT}\n"
}

func_run_cmd() {
    args=("$@")
    ELEMENTS=${#args[@]}

    printf "${RED}\$ "
    for (( i=0; i<$ELEMENTS; i++ )); do
        printf "%s " ${args[${i}]}
    done

    printf "${WIT}\n"
    eval $@
}

func_kill_sad() {
    ps axf | grep $SADProcess | grep -v grep | awk '{print "kill -9 " $1}' | sh
    func_print "Subject Access Delegation controller killed."
}

func_test_installed() {
    if ! command -v $1 >> /dev/null ; then
        func_print "$1 is not installed, exiting..."
        exit 1
    fi
}

make build_$OS

func_test_installed "minikube"
func_test_installed "kubectl"

func_print "-- Running end-to-end testing. --"

func_print "Deleting current minikube for testing..."
func_run_cmd $CmdDelete

func_print "Starting minikube (2GB memory)..."
func_run_cmd "$CmdStart"
func_run_cmd $CmdSleep

func_print "Starting Subject Access Delegation Controller..."
func_run_cmd $CmdSAD

func_print "Pausing for spin up..."
func_run_cmd $CmdK8sGetAll

func_print "Adding testing Roles and Service Accounts..."
func_run_cmd $CmdCreateTestResources

func_print "Adding test rule [1]..."
func_run_cmd $CmdTest1

func_print "Killing Subject Access Delegation Controller..."
func_kill_sad

func_print "Stopping and deleting minikube cluster..."
func_run_cmd $CmdDelete

exit 0
