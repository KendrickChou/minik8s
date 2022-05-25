package invoker

import (
	"context"

	"minik8s.com/minik8s/pkg/aqualake/podpoolmanager"
)

type Invoker interface {
	InvokeAction(ctx context.Context, env string, function string)
}

type invoker struct {
	podPool podpoolmanager.PodPoolManager
}

func NewInvoker() Invoker {
	ivk := &invoker{
		podPool: podpoolmanager.NewPodPoolManager(),
	}

	return ivk
}

func (ivk *invoker) InvokeAction(ctx context.Context, env string, function string) {

}
