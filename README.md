# Go GOMAXPROCS Replicator

This project provides a diagnostic tool to demonstrate the "start-and-stop" throttling behavior caused by high `GOMAXPROCS` values in containerized environments with strict CPU limits (CFS Quota).

## The Problem
In Kubernetes or OpenShift environments, a container might have a limit of, for example, 200ms of CPU time per 100ms period. If `GOMAXPROCS` is set to a high value (e.g., 64), the Go runtime spawns 64 OS threads. These threads can consume the entire 20ms quota in less than 1ms. 

The Linux kernel then halts all execution in the container for the remaining 99ms of the period. This results in significantly inflated execution times for bursty workloads compared to a correctly tuned `GOMAXPROCS` (e.g., 1 or 2).

### Run 1: Default Behavior (Untuned)
```bash
podman run --rm -it --cpus="0.2" quay.io/midu/gomaxprocs-replicator:latest
```
* **Host CPUs detected:** 32
* **GOMAXPROCS set to:** 32
* **Execution Time:** **4.78 seconds**

### Run 2: Tuned Behavior (`GOMAXPROCS=2`)
```bash
podman run --rm -it -e GOMAXPROCS=2 --cpus="0.2" quay.io/midu/gomaxprocs-replicator:latest
```
* **Host CPUs detected:** 32
* **GOMAXPROCS set to:** 2
* **Execution Time:** **11.19 milliseconds**

### Run 3: Default Behaviour (Untuned) with go-1.25
```bash
podman run --rm -it --cpus="0.2" quay.io/midu/gomaxprocs-replicator:latest-1.25
=== GOMAXPROCS Replicator ===
Go version: go1.25.14
Host Logical CPUs (runtime.NumCPU): 32
Default GOMAXPROCS: 2
=====================================
Starting CPU burst simulation...
Burst completed in: 86.316936ms
If running under a tight K8s CPU limit, this time will be artificially inflated due to severe CFS throttling.
Container is now idling. Press Ctrl+C to exit.
```

## What Do These Results Prove?

By simply telling the Go runtime to use 2 OS threads instead of 32, the exact same mathematical workload completed **over 420 times faster** (from 4,780ms down to 11ms). 

This proves three critical aspects of how Go interacts with Linux cgroups (container limits):

### 1. Go is blind to Container CPU Limits (Pre-1.25)
When the Go runtime starts, it queries the operating system for the number of available CPUs to determine how many OS threads to spawn for its scheduler. In containers, Go reads the physical host's core count (32 cores) rather than the container's cgroup CPU quota (0.2 cores). 

### 2. The Linux CFS Scheduler Punishes Over-Threading
When a container is limited to `0.2` CPUs, the Linux Completely Fair Scheduler (CFS) allocates it 20 milliseconds of execution time per 100ms window. 
* In **Run 1**, Go spawned 32 threads. Those 32 threads immediately woke up, fought each other for CPU time, generated massive context-switching overhead, and burned through the 20ms quota in a fraction of a millisecond. The kernel then hard-paused the entire container for the remaining ~80ms of the window. This cycle repeated for nearly 5 seconds.
* In **Run 2**, Go spawned only 2 threads. These threads executed smoothly and efficiently, completing the workload well within a single 20ms scheduling window without thrashing or hitting the CFS penalty box.

### 3. Proper Sizing is Mandatory in Kubernetes
If you run Go applications in OpenShift or Kubernetes with CPU `limits` applied, and you do not restrict `GOMAXPROCS`, your application will suffer from severe latency spikes, poor throughput, and inefficient garbage collection.

## Solutions

To fix this issue in production environments:

- Follow the steps witin https://access.redhat.com/solutions/7130704

