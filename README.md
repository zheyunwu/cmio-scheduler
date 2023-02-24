# CMIO-scheduler

CMIO-scheduler is a Kubernetes scheduler based on the [*Scheduling Framework*](https://kubernetes.io/docs/concepts/scheduling-eviction/scheduling-framework/) that schedules tasks according to nodes' CPU, Memory and I/O speed (CMIO).

## Description

This project is the work of the master thesis "[An I/O-aware scheduler for containerized data-intensive HPC tasks in Kubernetes-based heterogeneous clusters](http://kth.diva-portal.org/smash/record.jsf?pid=diva2%3A1725008&dswid=5071)".

## Author

Zheyun Wu @ KTH Royal Institute of Technology

## Usage

### Build CMIO-scheduler

```bash
cd cmio-scheduler

go build
```

### Prepare the Docker Image

```bash
docker build -t localhost:30002/cmio-scheduler:0.0.1 .

docker push localhost:30002/cmio-scheduler:0.0.1
```

### Deploy

```bash
cd  deploy

kubectl apply -f rbac.yaml
kubectl apply -f configmap.yaml
kubectl apply -f deploy.yaml
```