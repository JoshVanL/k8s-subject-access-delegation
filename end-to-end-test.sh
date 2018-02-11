#!/bin/bash

RED='\e[1;31m'
BLU='\e[0;36m'
WIT='\e[0;0m'

uname=$(uname)
OS=${uname,,}

SADProcess="k8s_subject_access_delegation_${OS}_amd64"

CmdDelete="minikube delete"
CmdStart="minikube start --extra-config=apiserver.Authorization.Mode=RBAC --memory 3072"
CmdSAD="./${SADProcess} &"
CmdK8sGetAll="kubectl get all"
CmdSleep="sleep 10"


func_print() {
    printf "\n> ${BLU}$1${WIT}\n"
}

func_run_cmd() {
    args=("$@")
    ELEMENTS=${#args[@]}

    printf "${RED}\$ "
    for (( i=0; i<$ELEMENTS; i++ )); do
        printf "${args[${i}]} "
    done

    printf "${WIT}\n"
    $@
}

func_kill_sad() {
    ps axf | grep $SADProcess | grep -v grep | awk '{print "kill -9 " $1}' | sh
    func_print "Subject Access Delegation controller killed."
}

func_print "-- Running end-to-end testing. --"

#make build_$OS

if ! command -v minikube >> /dev/null ; then
    func_print "minikube is not installed, exiting..."
    exit 1
fi

func_print "Deleting current minikube for testing..."
func_run_cmd $CmdDelete

func_print "Starting minikube (3GB memory)..."
func_run_cmd "$CmdStart"
func_run_cmd $CmdSleep

func_print "Starting Subject Access Delegation Controller..."
func_run_cmd $CmdSAD

func_print "Pausing for spin up..."
func_run_cmd $CmdK8sGetAll

func_print "Killing Subject Access Delegation Controller..."
func_kill_sad

func_print "Stopping and deleting: minikube."
func_run_cmd $CmdDelete
