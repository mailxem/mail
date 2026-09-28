#!/bin/bash
set -e

export AWS_REGION=us-east-2
CLUSTER_NAME=posthoot.k8s.local
KOPS_STATE_STORE=s3://posthoot-kops-state-696777127513

echo "🚀 Kops cluster - x86 spot only (us-east-2)"

# S3 bucket (already created with versioning enabled)

export KOPS_STATE_STORE=$KOPS_STATE_STORE

echo "Creating cluster..."
kops create cluster \
  --name=$CLUSTER_NAME \
  --cloud=aws \
  --zones=us-east-2a,us-east-2b,us-east-2c \
  --control-plane-zones=us-east-2a \
  --node-count=2 \
  --control-plane-size=t3.small \
  --control-plane-volume-size=10 \
  --node-volume-size=16 \
  --kubernetes-version=1.31.0 \
  --networking=amazonvpc \
  --topology=public \
  --ssh-public-key=~/.ssh/id_ed25519.pub \
  --yes

echo "Configuring x86 spot instances..."
sleep 30

# Get instance group and add spot config
kops get ig nodes-us-east-2a -o yaml > /tmp/nodes-ig.yaml

cat > /tmp/nodes-spot-x86.yaml << 'EOF'
apiVersion: kops.k8s.io/v1alpha2
kind: InstanceGroup
metadata:
  name: nodes-us-east-2a
  labels:
    kops.k8s.io/cluster: posthoot.k8s.local
spec:
  role: Node
  machineType: m5.large
  maxSize: 10
  minSize: 1
  maxPrice: "0.05"
  mixedInstancesPolicy:
    instances:
    - m5.large
    - m5a.large
    - m5n.large
    - m6i.large
    - c5.large
    - c5a.large
    - c6i.large
    - r5.large
    onDemandBase: 0
    onDemandAboveBase: 0
    spotInstancePools: 8
  subnets:
  - us-east-2a
  - us-east-2b
  - us-east-2c
  rootVolumeSize: 16
  rootVolumeType: gp3
EOF

kops replace -f /tmp/nodes-spot-x86.yaml
kops update cluster $CLUSTER_NAME --yes

echo "Waiting for cluster..."
sleep 90
kops validate cluster --wait 15m

echo "✅ Done!"
kubectl get nodes -o wide
