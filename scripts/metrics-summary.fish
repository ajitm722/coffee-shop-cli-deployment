#!/usr/bin/env fish
# Prints Prometheus targets and a one-shot summary (requests, rps, p90)
# Defaults assume Minikube NodePorts:
#   - coffee-api         :30090
#   - coffee-prometheus  :30091

# --- Config (override by exporting before running) ---
if not set -q MINIKUBE_IP
  set -x MINIKUBE_IP (minikube ip)
end
if not set -q API_BASE
  set -x API_BASE http://$MINIKUBE_IP:30090/v1
end
if not set -q PROM_URL
  set -x PROM_URL http://$MINIKUBE_IP:30091
end
if not set -q WIN
  set -x WIN 5m
end
# Optional: set TRAFFIC=1 to generate 30 hits first, or set COUNT yourself
if not set -q COUNT
  set -x COUNT 30
end

echo "== Config =="
echo "MINIKUBE_IP: $MINIKUBE_IP"
echo "API_BASE   : $API_BASE"
echo "PROM_URL   : $PROM_URL"
echo "WIN        : $WIN"
echo ""

# --- Optional traffic to /v1/menu to ensure non-empty metrics ---
if set -q TRAFFIC
  echo "Generating $COUNT requests to $API_BASE/menu ..."
  for i in (seq 1 $COUNT)
    curl -s $API_BASE/menu > /dev/null
  end
  echo "Traffic done."
  # small wait for Prometheus to scrape
  sleep 5
end

# --- Targets (like make prom-targets) ---
echo "== Prometheus targets @ $PROM_URL =="
if not curl -fsS "$PROM_URL/-/healthy" > /dev/null
  echo "  ERROR: Prometheus not reachable on $PROM_URL"
  exit 1
end
curl -fsS "$PROM_URL/api/v1/targets" \
| jq -r '.data.activeTargets[]? | "\(.labels.job) @ \(.labels.instance): \(.health)\t\(.lastError)"'
echo ""

# --- Summary (like make metrics-summary) ---
echo "== Metrics Summary (window=$WIN) =="

# total requests in the last $WIN
set REQ (curl -fsSG $PROM_URL/api/v1/query --data-urlencode "query=sum(increase(coffee_menu_requests_total[$WIN]))" \
  | jq -r 'if (.data.result|length)>0 then .data.result[0].value[1] else "0" end')

# requests/sec over $WIN
set RPS (curl -fsSG $PROM_URL/api/v1/query --data-urlencode "query=sum(rate(coffee_menu_requests_total[$WIN]))" \
  | jq -r 'if (.data.result|length)>0 then .data.result[0].value[1] else "0" end')

# p90 latency (ms) over $WIN
set P90 (curl -fsSG $PROM_URL/api/v1/query --data-urlencode "query=histogram_quantile(0.90, sum(rate(coffee_menu_latency_seconds_bucket[$WIN])) by (le))" \
  | jq -r 'if (.data.result|length)>0 then (.data.result[0].value[1]|try (tonumber*1000) catch "NaN") else "NaN" end')

printf "• requests: %s in %s\n" $REQ $WIN
printf "• rps     : %s\n" $RPS
printf "• p90     : %s ms\n" $P90

