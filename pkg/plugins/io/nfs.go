package io

import (
	"context"
	"fmt"

	"k8s.io/klog"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubernetes/pkg/scheduler/framework"

	"net/http"
	"io/ioutil"
	"encoding/json"
	"strconv"
	"math"
)

// NFS is a score plugin
type NFS struct {
	handle framework.Handle
}

var _ framework.ScorePlugin = &NFS{}

// Name is the name of the plugin used in the plugin registry and configurations.
const Name = "NFS"

// Name returns name of the plugin. It is used in logs, etc.
func (pl *NFS) Name() string {
	return Name
}

// Score invoked at the score extension point.
func (pl *NFS) Score(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
	fmt.Println("[NFS Plugin] SCORE started")

	// Overall score
	var score int64 = 0

	// Query WRITE ops
	var wops float64 = queryNfsRate("Write", nodeName)
	score += int64(math.Round(wops * 10000))

	// Query READ ops
	var rops float64 = queryNfsRate("Read", nodeName)
	score += int64(math.Round(rops * 10000))

	fmt.Println("[NFS Plugin] SCORE finished ", nodeName, score)
	return score, nil
}

func (pl *NFS) NormalizeScore(ctx context.Context, state *framework.CycleState, pod *v1.Pod, scores framework.NodeScoreList) *framework.Status {
	var (
		highest int64 = 0
		lowest        = scores[0].Score
	)

	for _, nodeScore := range scores {
		if nodeScore.Score < lowest {
			lowest = nodeScore.Score
		}
		if nodeScore.Score > highest {
			highest = nodeScore.Score
		}
	}

	if highest == lowest {
		lowest--
	}

	// Normalize scores to the range [0-100]
	for i, nodeScore := range scores {
		scores[i].Score = (nodeScore.Score - lowest) * framework.MaxNodeScore / (highest - lowest)
		klog.Infof("[NFS Plugin] Node: %v, Score: %v When scheduling Pod: %v/%v", scores[i].Name, scores[i].Score, pod.GetNamespace(), pod.GetName())
	}

	klog.Infof("[NFS Plugin] Nodes final score: %v", scores)
	return nil
}

// ScoreExtensions of the Score plugin.
func (pl *NFS) ScoreExtensions() framework.ScoreExtensions {
	return pl
}

// New initializes a new plugin and returns it.
func New(_ runtime.Object, h framework.Handle) (framework.Plugin, error) {
	return &NFS{handle: h}, nil
}

type Res struct {
	Data Data
	Status string
}

type Data struct {
	Result []ResultItem
	ResultType string
}

type ResultItem struct {
	Metric map[string]any
	Value []interface{}
}

func queryNfsRate(method string, nodeName string) (float64) {
	// Given a NFS method (Write or Read), return operations per second
	var rate float64 = 0.0

	url := "http://kube-prometheus-stack-1660-prometheus.prometheus:9090/api/v1/query?query=rate(node_nfs_requests_total{method='" + method + "',+proto='4'}[7d])+*+on(instance)+group_left(nodename)+(node_uname_info{nodename=~'(?i:(" + nodeName + "))'})"
	response, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
	}

	var res Res
	if err := json.Unmarshal([]byte(string(body)), &res); err != nil {
		fmt.Println(err)
	}

	fmt.Println("Parsed JSON", nodeName, res.Data)

	if len(res.Data.Result) > 0 {
		ops_str, _ := (res.Data.Result[0].Value[1]).(string)
		ops, _ := strconv.ParseFloat(ops_str, 64)
		rate = ops
	}

	return rate
}