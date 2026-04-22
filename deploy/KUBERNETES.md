# Kubernetes Deployment Guide

This guide covers deploying the SNID Benchmarking Platform to Kubernetes.

## Prerequisites

- Kubernetes cluster (v1.24+)
- kubectl configured
- 4GB+ RAM per node
- StorageClass for persistent volumes

## Quick Start

### 1. Create Namespace

```bash
kubectl create namespace snid-benchmarks
```

### 2. Create Persistent Volume Claim

```yaml
# pvc.yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: benchmark-results
  namespace: snid-benchmarks
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
  storageClassName: standard
```

```bash
kubectl apply -f pvc.yaml
```

### 3. Create Deployment

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: snid-benchmarks
  namespace: snid-benchmarks
  labels:
    app: snid-benchmarks
spec:
  replicas: 1
  selector:
    matchLabels:
      app: snid-benchmarks
  template:
    metadata:
      labels:
        app: snid-benchmarks
    spec:
      containers:
      - name: snid-benchmarks
        image: snid-benchmarks:latest
        imagePullPolicy: IfNotPresent
        ports:
        - containerPort: 8080
          name: http
        env:
        - name: BENCH_MODE
          value: "web"
        - name: BENCH_PURE_MODE
          value: "true"
        - name: RESULTS_DIR
          value: "/app/results"
        - name: PORT
          value: "8080"
        - name: REGRESSION_THRESHOLD
          value: "10"
        volumeMounts:
        - name: results
          mountPath: /app/results
        resources:
          requests:
            memory: "2Gi"
            cpu: "1"
          limits:
            memory: "4Gi"
            cpu: "2"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
      volumes:
      - name: results
        persistentVolumeClaim:
          claimName: benchmark-results
```

```bash
kubectl apply -f deployment.yaml
```

### 4. Create Service

```yaml
# service.yaml
apiVersion: v1
kind: Service
metadata:
  name: snid-benchmarks
  namespace: snid-benchmarks
spec:
  selector:
    app: snid-benchmarks
  ports:
  - port: 80
    targetPort: 8080
    name: http
  type: LoadBalancer
```

```bash
kubectl apply -f service.yaml
```

### 5. Access Dashboard

```bash
# Get external IP
kubectl get service snid-benchmarks -n snid-benchmarks

# Or use port-forward for local access
kubectl port-forward service/snid-benchmarks 8080:80 -n snid-benchmarks
```

## CLI Mode Job

For one-off benchmark runs:

```yaml
# benchmark-job.yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: snid-benchmark-run
  namespace: snid-benchmarks
spec:
  template:
    spec:
      containers:
      - name: benchmark
        image: snid-benchmarks:latest
        env:
        - name: BENCH_MODE
          value: "cli"
        - name: BENCH_PURE_MODE
          value: "true"
        - name: BENCH_SUITES
          value: "all"
        - name: RESULTS_DIR
          value: "/app/results"
        volumeMounts:
        - name: results
          mountPath: /app/results
      volumes:
      - name: results
        persistentVolumeClaim:
          claimName: benchmark-results
      restartPolicy: Never
```

```bash
kubectl apply -f benchmark-job.yaml
```

## CronJob for Scheduled Runs

```yaml
# cronjob.yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: snid-nightly-benchmarks
  namespace: snid-benchmarks
spec:
  schedule: "0 0 * * *"  # Daily at midnight UTC
  successfulJobsHistoryLimit: 3
  failedJobsHistoryLimit: 1
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: benchmark
            image: snid-benchmarks:latest
            env:
            - name: BENCH_MODE
              value: "cli"
            - name: BENCH_SUITES
              value: "all"
            volumeMounts:
            - name: results
              mountPath: /app/results
          volumes:
          - name: results
            persistentVolumeClaim:
              claimName: benchmark-results
          restartPolicy: OnFailure
```

```bash
kubectl apply -f cronjob.yaml
```

## Horizontal Pod Autoscaling

```yaml
# hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: snid-benchmarks-hpa
  namespace: snid-benchmarks
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: snid-benchmarks
  minReplicas: 1
  maxReplicas: 5
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

```bash
kubectl apply -f hpa.yaml
```

## Ingress Configuration

For custom domain and TLS:

```yaml
# ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: snid-benchmarks-ingress
  namespace: snid-benchmarks
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
spec:
  tls:
  - hosts:
    - benchmarks.example.com
    secretName: snid-benchmarks-tls
  rules:
  - host: benchmarks.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: snid-benchmarks
            port:
              number: 80
```

## Monitoring and Logging

### Prometheus Metrics

Add Prometheus annotations to deployment:

```yaml
annotations:
  prometheus.io/scrape: "true"
  prometheus.io/port: "8080"
  prometheus.io/path: "/metrics"
```

### Log Aggregation

```bash
# View logs
kubectl logs -f deployment/snid-benchmarks -n snid-benchmarks

# View logs for specific pod
kubectl logs -f <pod-name> -n snid-benchmarks
```

## Troubleshooting

### Pod Not Starting

```bash
# Check pod status
kubectl get pods -n snid-benchmarks

# Describe pod for events
kubectl describe pod <pod-name> -n snid-benchmarks

# View logs
kubectl logs <pod-name> -n snid-benchmarks
```

### PVC Issues

```bash
# Check PVC status
kubectl get pvc -n snid-benchmarks

# Describe PVC
kubectl describe pvc benchmark-results -n snid-benchmarks
```

### Resource Limits

Adjust resource requests/limits in deployment.yaml based on cluster capacity.

## Scaling

### Manual Scaling

```bash
kubectl scale deployment snid-benchmarks --replicas=3 -n snid-benchmarks
```

### Auto Scaling

See HPA configuration above.

## Backup and Restore

### Backup Results

```bash
# Copy results from PVC
kubectl run -it --rm backup \
  --image=busybox \
  -n snid-benchmarks \
  --overrides='
{
  "spec": {
    "containers": [{
      "name": "backup",
      "image": "busybox",
      "command": ["tar", "czf", "/backup/results.tar.gz", "/app/results"],
      "volumeMounts": [{
        "name": "results",
        "mountPath": "/app/results"
      }, {
        "name": "backup",
        "mountPath": "/backup"
      }]
    }],
    "volumes": [{
      "name": "results",
      "persistentVolumeClaim": {"claimName": "benchmark-results"}
    }, {
      "name": "backup",
      "persistentVolumeClaim": {"claimName": "backup-pvc"}
    }]
  }
}'

kubectl cp backup-pvc:/backup/results.tar.gz ./results-backup.tar.gz
```

## Upgrading

```bash
# Build new image
docker build -f benchmarks/Dockerfile -t snid-benchmarks:v2.0.0 .

# Push to registry
docker tag snid-benchmarks:v2.0.0 your-registry/snid-benchmarks:v2.0.0
docker push your-registry/snid-benchmarks:v2.0.0

# Update deployment image
kubectl set image deployment/snid-benchmarks \
  snid-benchmarks=your-registry/snid-benchmarks:v2.0.0 \
  -n snid-benchmarks

# Rollout status
kubectl rollout status deployment/snid-benchmarks -n snid-benchmarks
```

## Cleanup

```bash
# Delete all resources
kubectl delete namespace snid-benchmarks

# Or delete individual resources
kubectl delete deployment snid-benchmarks -n snid-benchmarks
kubectl delete service snid-benchmarks -n snid-benchmarks
kubectl delete pvc benchmark-results -n snid-benchmarks
```
