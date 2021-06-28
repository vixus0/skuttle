#!/usr/bin/env bash

set -o errexit -o nounset -o pipefail

tmpd="$(mktemp -d -p /tmp skuttle.XXXXX)"
echo "config dir: $tmpd"

export KUBECONFIG="$tmpd/config"
export NODE_LIST="$tmpd/nodes"

# set up provider nodes
cat <<EOF >"$NODE_LIST"
node1
node3
EOF

# create kind cluster
kind delete cluster --name skuttle
kind create cluster \
  --config integration/kind.yaml \
  --name skuttle \
  --wait 2m

# start skuttle
./skuttle \
  --log-level debug \
  --node-selector "!node-role.kubernetes.io/control-plane" \
  --not-ready-duration 20s \
  --providers file \
  >"$tmpd/log" 2>&1 &

skuttle_pid="$!"

# follow logs
tail -f "$tmpd/log" &

# make a worker not ready
docker exec skuttle-worker2 systemctl stop kubelet

# wait for worker to be removed by skuttle
exitcode=0
waits=0
while kubectl get node skuttle-worker2 >/dev/null 2>/dev/null; do
  sleep 1
  waits="$(( waits+1 ))"

  if test "$waits" -eq 30; then
    echo "FAIL: timed out waiting for node deletion"
    exitcode=1
  fi
done

echo "PASS: node deleted by skuttle"

# cleanup
kill "$skuttle_pid"
kind delete cluster --name skuttle
rm -r "$tmpd"

# exit
exit $exitcode
