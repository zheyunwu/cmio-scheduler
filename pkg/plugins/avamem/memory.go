package avamem

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

// Memory is a score plugin
type Memory struct {
	handle framework.Handle
}

var _ framework.ScorePlugin = &Memory{}

// Name is the name of the plugin used in the plugin registry and configurations.
const Name = "AvaMem"

// Name returns name of the plugin. It is used in logs, etc.
func (pl *Memory) Name() string {
	return Name
}

// Score invoked at the score extension point.
func (pl *Memory) Score(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
	fmt.Println("[Memory Plugin] SCORE started")

	// Overall score
	var score int64 = 0

	// Query avaiable memory
	var ava_mem float64 = queryAvaMem(nodeName)
	score += int64(math.Round(ava_mem))

	fmt.Println("[Memory Plugin] SCORE finished ", nodeName, score)
	return score, nil
}

func (pl *Memory) NormalizeScore(ctx context.Context, state *framework.CycleState, pod *v1.Pod, scores framework.NodeScoreList) *framework.Status {
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
		klog.Infof("[Memory Plugin] Node: %v, Score: %v When scheduling Pod: %v/%v", scores[i].Name, scores[i].Score, pod.GetNamespace(), pod.GetName())
	}

	klog.Infof("[Memory Plugin] Nodes final score: %v", scores)
	return nil
}

// ScoreExtensions of the Score plugin.
func (pl *Memory) ScoreExtensions() framework.ScoreExtensions {
	return pl
}

// New initializes a new plugin and returns it.
func New(_ runtime.Object, h framework.Handle) (framework.Plugin, error) {
	return &Memory{handle: h}, nil
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

func queryAvaMem(nodeName string) (float64) {
	// Return avaiable memory in MB
	var memory float64 = 0.0

	url := "http://kube-prometheus-stack-1660-prometheus.prometheus:9090/api/v1/query?query=node_memory_MemAvailable_bytes/1024/1024+*+on(instance)+group_left(nodename)+(node_uname_info{nodename=~'(?i:(" + nodeName + "))'})"
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
		value_str, _ := (res.Data.Result[0].Value[1]).(string)
		value, _ := strconv.ParseFloat(value_str, 64)
		memory = value
	}

	return memory
}