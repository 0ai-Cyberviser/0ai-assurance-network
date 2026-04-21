"""Multi-model routing engine for governance inference."""

from __future__ import annotations

import json
from dataclasses import dataclass
from pathlib import Path
from typing import Any

from .config import LoadedConfig
from .model_registry import ModelRegistry, get_default_registry


@dataclass(frozen=True)
class ModelRoutingDecision:
    """Decision about which models to use for inference."""

    strategy: str
    selected_models: list[str]
    fallback_chain: list[str]
    rationale: str
    max_latency_ms: int
    requires_consensus: bool


@dataclass(frozen=True)
class MultiModelInferenceResult:
    """Result from multi-model inference."""

    proposal_id: str
    title: str
    routing_decision: ModelRoutingDecision
    model_results: dict[str, dict[str, Any]]
    aggregated_result: dict[str, Any]
    consensus_achieved: bool
    confidence: float
    execution_time_ms: float
    summary: str


def load_routing_policy(config: LoadedConfig) -> dict[str, Any]:
    """Load model routing policy configuration."""
    policy_path = config.root / "config" / "governance" / "routing-policy.json"
    if not policy_path.exists():
        return {"default_strategy": "waterfall"}
    with policy_path.open("r", encoding="utf-8") as handle:
        return json.load(handle)


def select_routing_strategy(
    proposal: dict[str, Any],
    proposal_class: str,
    policy: dict[str, Any],
) -> ModelRoutingDecision:
    """Select appropriate routing strategy based on proposal characteristics."""
    selection_rules = policy.get("model_selection_rules", {})
    strategies = policy.get("strategies", {})
    fallback_chains = policy.get("fallback_chains", {})
    default_strategy = policy.get("default_strategy", "waterfall")

    # Determine which rule applies
    rule_key = None
    if proposal.get("is_emergency"):
        rule_key = "emergency_proposals"
    elif proposal_class == "safety_critical":
        rule_key = "safety_critical_proposals"
    elif proposal_class == "high_impact":
        rule_key = "high_impact_proposals"
    else:
        rule_key = "standard_proposals"

    rule = selection_rules.get(rule_key, {})
    strategy = rule.get("strategy", default_strategy)
    required_models = rule.get("required_models", ["deterministic"])
    optional_models = rule.get("optional_models", [])
    max_latency = rule.get("max_latency_ms", 10000)
    min_confidence = rule.get("min_confidence", 0.65)

    # Get fallback chain
    fallback_chain = fallback_chains.get(
        "high_confidence" if min_confidence > 0.75 else "default",
        ["deterministic"]
    )

    selected_models = list(required_models)
    strategy_config = strategies.get(strategy, {})
    requires_consensus = strategy == "consensus"

    if strategy == "consensus":
        # Add optional models for consensus
        selected_models.extend(optional_models)
        # Ensure minimum models for consensus
        min_models = strategy_config.get("min_models", 2)
        if len(selected_models) < min_models:
            selected_models.extend(
                [m for m in fallback_chain if m not in selected_models][:min_models]
            )

    rationale = (
        f"Selected {strategy} strategy for {rule_key} "
        f"with {len(selected_models)} model(s): {', '.join(selected_models)}"
    )

    return ModelRoutingDecision(
        strategy=strategy,
        selected_models=selected_models,
        fallback_chain=fallback_chain,
        rationale=rationale,
        max_latency_ms=max_latency,
        requires_consensus=requires_consensus,
    )


def aggregate_model_results(
    model_results: dict[str, dict[str, Any]],
    strategy: str,
    policy: dict[str, Any],
) -> tuple[dict[str, Any], bool, float]:
    """Aggregate results from multiple models.

    Returns: (aggregated_result, consensus_achieved, confidence)
    """
    if not model_results:
        return {}, False, 0.0

    aggregation_rules = policy.get("aggregation_rules", {})
    strategy_config = aggregation_rules.get(strategy, {})

    if strategy == "waterfall":
        # Use the first successful result
        for model_id in model_results:
            result = model_results[model_id]
            if result.get("success", False):
                return result, True, result.get("confidence", 0.5)
        return {}, False, 0.0

    elif strategy == "consensus":
        # Weighted voting across models
        method = strategy_config.get("method", "weighted_voting")
        conflict_resolution = strategy_config.get("conflict_resolution", "most_conservative")

        # Collect all results
        classes = []
        risk_scores = []
        confidences = []
        all_signals = []
        all_rationale = []
        all_remediation = []

        for model_id, result in model_results.items():
            if result.get("success", False):
                classes.append(result.get("proposal_class", "standard"))
                risk_scores.append(result.get("risk_score", 0))
                confidences.append(result.get("confidence", 0.5))
                all_signals.extend(result.get("triggered_signals", []))
                all_rationale.extend(result.get("rationale", []))
                all_remediation.extend(result.get("remediation", []))

        if not classes:
            return {}, False, 0.0

        # Use most conservative classification
        class_rank = {"safety_critical": 2, "high_impact": 1, "standard": 0}
        aggregated_class = max(classes, key=lambda c: class_rank.get(c, 0))

        # Average risk scores and confidence
        avg_risk_score = int(sum(risk_scores) / len(risk_scores))
        avg_confidence = sum(confidences) / len(confidences)

        # Deduplicate signals, rationale, and remediation
        unique_signals = list(dict.fromkeys(all_signals))
        unique_rationale = list(dict.fromkeys(all_rationale))
        unique_remediation = list(dict.fromkeys(all_remediation))

        # Check consensus
        consensus_threshold = policy.get("strategies", {}).get("consensus", {}).get("consensus_threshold", 0.67)
        most_common_class = max(set(classes), key=classes.count)
        consensus_count = classes.count(most_common_class)
        consensus_achieved = (consensus_count / len(classes)) >= consensus_threshold

        aggregated = {
            "success": True,
            "proposal_class": aggregated_class,
            "risk_score": avg_risk_score,
            "confidence": round(avg_confidence, 2),
            "triggered_signals": unique_signals,
            "rationale": unique_rationale,
            "remediation": unique_remediation,
            "model_count": len(model_results),
            "consensus_class": most_common_class,
        }

        return aggregated, consensus_achieved, avg_confidence

    elif strategy == "hybrid":
        # Layered enhancement approach
        base_weight = strategy_config.get("base_weight", 0.7)
        enhancement_weight = strategy_config.get("enhancement_weight", 0.3)

        # Find base model result (should be deterministic)
        base_result = model_results.get("deterministic", {})
        if not base_result.get("success", False):
            return {}, False, 0.0

        # Collect enhancement results
        enhancement_signals = []
        enhancement_remediation = []
        for model_id, result in model_results.items():
            if model_id != "deterministic" and result.get("success", False):
                enhancement_signals.extend(result.get("triggered_signals", []))
                enhancement_remediation.extend(result.get("remediation", []))

        # Merge base with enhancements
        aggregated = dict(base_result)
        aggregated["triggered_signals"] = list(dict.fromkeys(
            base_result.get("triggered_signals", []) + enhancement_signals
        ))
        aggregated["remediation"] = list(dict.fromkeys(
            base_result.get("remediation", []) + enhancement_remediation
        ))

        # Adjust confidence based on enhancement agreement
        base_confidence = base_result.get("confidence", 0.5)
        adjusted_confidence = base_confidence * base_weight + (0.8 * enhancement_weight)
        aggregated["confidence"] = round(adjusted_confidence, 2)

        return aggregated, True, adjusted_confidence

    else:
        # Default: use specialized model if available
        for model_id in model_results:
            result = model_results[model_id]
            if result.get("success", False):
                return result, True, result.get("confidence", 0.5)
        return {}, False, 0.0


def route_inference(
    config: LoadedConfig,
    proposal: dict[str, Any],
    proposal_class: str,
    *,
    policy: dict[str, Any] | None = None,
    registry: ModelRegistry | None = None,
) -> MultiModelInferenceResult:
    """Route inference request to appropriate models."""
    import time

    start_time = time.time()
    routing_policy = policy or load_routing_policy(config)
    model_registry = registry or get_default_registry()

    proposal_id = str(proposal.get("proposal_id", "unknown"))
    title = str(proposal.get("title", "Untitled proposal"))

    # Select routing strategy
    routing_decision = select_routing_strategy(
        proposal,
        proposal_class,
        routing_policy
    )

    # Execute models
    model_results = {}
    for model_id in routing_decision.selected_models:
        model = model_registry.get_model(model_id)
        if model is None:
            model_results[model_id] = {
                "success": False,
                "error": "Model not available"
            }
            continue

        try:
            result = model.infer(config, proposal)
            model_results[model_id] = {
                "success": True,
                **result
            }
        except Exception as e:
            model_results[model_id] = {
                "success": False,
                "error": str(e)
            }

    # Aggregate results
    aggregated, consensus, confidence = aggregate_model_results(
        model_results,
        routing_decision.strategy,
        routing_policy
    )

    execution_time = (time.time() - start_time) * 1000

    summary = (
        f"Multi-model inference using {routing_decision.strategy} strategy: "
        f"{len([r for r in model_results.values() if r.get('success')])} of "
        f"{len(model_results)} models succeeded"
    )
    if routing_decision.requires_consensus:
        summary += f", consensus {'achieved' if consensus else 'not achieved'}"

    return MultiModelInferenceResult(
        proposal_id=proposal_id,
        title=title,
        routing_decision=routing_decision,
        model_results=model_results,
        aggregated_result=aggregated,
        consensus_achieved=consensus,
        confidence=confidence,
        execution_time_ms=round(execution_time, 2),
        summary=summary,
    )
