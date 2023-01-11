package flowservice

const (
	ConclusionGo            = 1
	ConclusionGoWithRisk    = 2
	ConclusionReject        = 3
	ConCLusionTransferOther = 100
	CurStepTblColName       = "cur_step"

	// 状态 1: 流程草稿   2: 流程流转中  3: 流程完成（被拒绝）， 4: 流程完成（超时） 5： 流程完成（正常）
	FlowStateDraft   = 1
	FlowStateRunning = 2
	FlowStateFinish  = 3
	FlowStateReject  = 4
	FlowStateTimeOut = 5
)
