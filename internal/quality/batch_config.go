package quality

func BuildBatchPlan(size int, stages []BatchStage) BatchPlan {
	return BatchPlan{Size: size, Stages: append([]BatchStage(nil), stages...)}
}
