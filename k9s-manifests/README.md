# Go GOMAXPROCS Replicator

This project provides a diagnostic tool to demonstrate the "start-and-stop" throttling behavior caused by high `GOMAXPROCS` values in containerized environments with strict CPU limits (CFS Quota).

## The Problem
In Kubernetes or OpenShift environments, a container might have a limit of, for example, 200ms of CPU time per 100ms period. If `GOMAXPROCS` is set to a high value (e.g., 64), the Go runtime spawns 64 OS threads. These threads can consume the entire 20ms quota in less than 1ms. 

The Linux kernel then halts all execution in the container for the remaining 99ms of the period. This results in significantly inflated execution times for bursty workloads compared to a correctly tuned `GOMAXPROCS` (e.g., 1 or 2).

## Kubernetes / OpenShift Deployment

To demonstrate this issue on a live cluster, you can deploy the replicator application using the provided `deployment.yaml` manifest. The manifest creates a dedicated namespace and enforces a `200m` CPU limit.

### 1. Apply the Manifest

Ensure you have updated the `image` field in `deployment.yaml` to point to your pushed image. Then, apply it to the cluster:

- Namespace:

```bash
kubectl apply -f namespace.yaml
# or
oc apply -f namespace.yaml
```

- Deployment:

```bash
kubectl apply -f deployment.yaml
# or
oc apply -f deployment.yaml
```

### 2. Verify the Pod is Running

Check the status of the pod in the newly created namespace:

```bash
kubectl get pods -n gomaxprocs-replicator
# or
oc get pods -n gomaxprocs-replicator
```

### 3. Observe the Logs (The Throttling Issue)

Once the pod is running, fetch the logs to see the execution time of the CPU burst. Assuming your worker node has 32 cores, you will see output similar to this:

```bash
kubectl logs -l app=gomaxprocs-replicator -n gomaxprocs-replicator
# or
oc logs -l app=gomaxprocs-replicator -n gomaxprocs-replicator
```

**Expected Output:**
```text
=== Go 1.24 GOMAXPROCS Replicator ===
Go version: go1.24.13
Host Logical CPUs (runtime.NumCPU): 32
Default GOMAXPROCS: 32
=====================================
Starting CPU burst simulation...
Burst completed in: 4.781024499s
If running under a tight K8s CPU limit, this time will be artificially inflated due to severe CFS throttling.
Container is now idling. Press Ctrl+C to exit.
```

*Note: The burst took nearly 5 seconds instead of milliseconds because the 32 Go threads instantly consumed the 20ms CPU quota, forcing the K8s node's kernel to freeze the container for the rest of every 100ms scheduling window.*

### 4. Fixing the Issue in Kubernetes

To fix this issue without upgrading to Go 1.25+, you can use the Kubernetes Downward API to dynamically inject the CPU limit as the `GOMAXPROCS` environment variable. Add this to your container spec in `deployment.yaml`:

```yaml
env:
  - name: GOMAXPROCS
    valueFrom:
      resourceFieldRef:
        resource: limits.cpu
        divisor: "1"
```
Redeploying with this environment variable will drop the burst time back down to milliseconds.

## Solutions

To fix this issue in production environments:

- Follow the steps witin https://access.redhat.com/solutions/7130704

