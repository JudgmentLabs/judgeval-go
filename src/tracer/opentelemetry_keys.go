package tracer

var AttributeKeys = struct {
	JudgmentSpanKind          string
	JudgmentInput             string
	JudgmentOutput            string
	JudgmentOfflineMode       string
	JudgmentUpdateID          string
	JudgmentCustomerID        string
	JudgmentAgentID           string
	JudgmentParentAgentID     string
	JudgmentAgentClassName    string
	JudgmentAgentInstanceName string
	JudgmentIsAgentEntryPoint string
	JudgmentCumulativeLLMCost string
	JudgmentStateBefore       string
	JudgmentStateAfter        string
	PendingTraceEval          string
}{
	JudgmentSpanKind:          "judgment.span_kind",
	JudgmentInput:             "judgment.input",
	JudgmentOutput:            "judgment.output",
	JudgmentOfflineMode:       "judgment.offline_mode",
	JudgmentUpdateID:          "judgment.update_id",
	JudgmentCustomerID:        "judgment.customer_id",
	JudgmentAgentID:           "judgment.agent_id",
	JudgmentParentAgentID:     "judgment.parent_agent_id",
	JudgmentAgentClassName:    "judgment.agent_class_name",
	JudgmentAgentInstanceName: "judgment.agent_instance_name",
	JudgmentIsAgentEntryPoint: "judgment.is_agent_entry_point",
	JudgmentCumulativeLLMCost: "judgment.cumulative_llm_cost",
	JudgmentStateBefore:       "judgment.state_before",
	JudgmentStateAfter:        "judgment.state_after",
	PendingTraceEval:          "judgment.pending_trace_eval",
}

var ResourceKeys = struct {
	ServiceName          string
	TelemetrySDKLanguage string
	TelemetrySDKName     string
	TelemetrySDKVersion  string
}{
	ServiceName:          "service.name",
	TelemetrySDKLanguage: "telemetry.sdk.language",
	TelemetrySDKName:     "telemetry.sdk.name",
	TelemetrySDKVersion:  "telemetry.sdk.version",
}
