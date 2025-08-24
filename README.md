# Kubernetes Custom Controller - Sidecar Shutdown

Kubernetes (cron)jobs sidecar terminator.

_Originally forked from https://github.com/nrmitchi/k8s-controller-sidecars_

> [!NOTE]
> Official Kubernetes [Sidecar Containers](https://kubernetes.io/docs/concepts/workloads/pods/sidecar-containers/) are now stable (v1.33, enabled by default from v1.29). This eliminates the need for this operator.
> Define the sidecar as a restartable init container (`initContainers[].restartPolicy: Always`). The Kubelet will keep it running during the Pod’s lifecycle and automatically stop it when the app container(s) finish.

## What is this?

This is a custom Kubernetes controller for the purpose of watching running pods, and sending a SIGTERM to sidecar containers when the "main" application container has exited (and the sidecars are the only non-terminated containers).

This is a response to https://github.com/kubernetes/kubernetes/issues/25908.

## Caveats

 The Kubernetes API has no supported “send signal to a single container” primitive other than `exec`. It's impossible to arbitrarily signal a single container inside a pod without a process running there to receive it. Therefore, this operator relies on `/bin/sh` being present in the sidecar container images to run `kill -s SIGTERM 1`.

 In some minimal containers, like Chainguard's secure base images, a shell is not included by default.

## Usage

1. Deploy the controller into your cluster

```sh
kubectl apply -f manifest.yml
```

1. Add the `pod.kubernetes.io/sidecars` annotation to your pods, with a comma-separated list of sidecar container names.

Example:

```yaml
---
apiVersion: batch/v1
kind: CronJob
metadata:
  name: test-job
spec:
  schedule: "* * * * *"
  jobTemplate:
    spec:
      template:
        metadata:
          annotations:
            pod.kubernetes.io/sidecars: cloud-sql-proxy
        spec:
          restartPolicy: Never
          containers:
            - name: test-job
              image: ubuntu:latest
              command: ["sleep", "15"]
            - name: cloud-sql-proxy
              image: gcr.io/cloudsql-docker/gce-proxy:1.33.0-alpine
              command: [ "/cloud_sql_proxy" ]
```

## Dependency Management

This project uses Go modules for dependency management.

To upgrade dependencies, use [go-mod-upgrade](https://github.com/oligot/go-mod-upgrade):

```sh
go install github.com/oligot/go-mod-upgrade@latest
go-mod-upgrade
```
