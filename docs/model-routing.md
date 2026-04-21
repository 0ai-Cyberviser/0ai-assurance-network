# Multi-Model Routing

## Overview

The 0AI Assurance Network supports multi-model routing for governance inference, enabling dynamic selection and composition of multiple inference engines based on proposal characteristics, confidence requirements, and performance constraints.

## Purpose

Multi-model routing provides:

- **Redundancy**: Fallback to alternative models when primary models are unavailable
- **Consensus**: Run multiple models and aggregate results to reduce single-point-of-failure risk
- **Specialization**: Route to domain-specific models based on proposal type
- **Enhancement**: Layer AI-backed analysis on top of deterministic base models
- **Flexibility**: Configure routing strategies per proposal class

## Routing Strategies

### Waterfall

Try models in priority order with automatic fallback:

- Attempts models sequentially until one succeeds
- Fast and efficient for most cases
- Gracefully degrades when models are unavailable
- **Use case**: Standard proposals, general-purpose inference

### Consensus

Run multiple models and aggregate results via voting:

- Executes 2+ models in parallel
- Aggregates via weighted voting or most-conservative selection
- Detects disagreements and conflicts
- **Use case**: Safety-critical proposals, high-stakes decisions
- **Configuration**: Requires `consensus_threshold` (e.g., 0.67 = 67% agreement)

### Specialized

Route to domain-specific models based on proposal characteristics:

- Maps proposal types to specialized models
- Falls back to general models if specialist unavailable
- **Use case**: Treasury grants → financial model, validator changes → infrastructure model

### Hybrid

Combine deterministic base with ML enhancement:

- Runs deterministic model as authoritative base (70% weight)
- Adds ML-backed analysis for additional signals (30% weight)
- Merges signals and remediation from both
- **Use case**: High-impact proposals needing both rigor and breadth

## Configuration

Multi-model routing is configured via `config/governance/routing-policy.json`:

```json
{
  "version": "routing-policy-2026-04-21",
  "default_strategy": "waterfall",
  "strategies": { ... },
  "model_selection_rules": {
    "safety_critical_proposals": {
      "strategy": "consensus",
      "required_models": ["deterministic", "threat-detection"],
      "min_confidence": 0.85
    }
  }
}
```

## Model Selection Rules

Routing decisions are made based on proposal classification:

| Proposal Class | Strategy | Required Models | Min Confidence |
|----------------|----------|-----------------|----------------|
| **Emergency** | Consensus | deterministic, threat-detection | 0.90 |
| **Safety Critical** | Consensus | deterministic, threat-detection | 0.85 |
| **High Impact** | Hybrid | deterministic | 0.75 |
| **Standard** | Waterfall | deterministic | 0.65 |

## Model Registry

The model registry manages available inference engines:

### Built-in Models

- **deterministic**: The existing rule-based governance inference engine (always available)
- **threat-detection**: Zero-day threat and vulnerability scanner
- **fallback-deterministic**: Simplified fallback when primary models fail

### Model Metadata

Each model provides:

- `model_id`: Unique identifier
- `model_version`: Version string
- `capabilities`: Supported features (e.g., "threat-detection", "proposal-classification")
- `confidence_range`: Expected confidence bounds
- `average_latency_ms`: Typical response time

### Health Checking

The registry tracks model health:

- Periodic health checks (default: every 60 seconds)
- Failure threshold (default: 3 consecutive failures)
- Recovery cooldown (default: 300 seconds)
- Automatic unavailability marking

## Usage

### Command Line

Run multi-model inference with automatic routing:

```bash
PYTHONPATH=src python -m assurancectl.cli governance-multi-model \
  --proposal examples/proposals/emergency-pause.json
```

Override routing strategy:

```bash
PYTHONPATH=src python -m assurancectl.cli governance-multi-model \
  --proposal examples/proposals/emergency-pause.json \
  --strategy consensus
```

Generate machine-readable output:

```bash
PYTHONPATH=src python -m assurancectl.cli governance-multi-model \
  --proposal examples/proposals/emergency-pause.json \
  --json \
  --artifact-out build/artifacts/multi-model.json
```

### Make Targets

```bash
make governance-multi-model PROPOSAL=examples/proposals/emergency-pause.json
```

## Result Aggregation

### Consensus Aggregation

When using consensus strategy:

1. Collects results from all models
2. Uses most conservative classification (safety_critical > high_impact > standard)
3. Averages risk scores and confidence
4. Deduplicates signals, rationale, and remediation
5. Checks consensus threshold (default: 67%)
6. Reports consensus achievement status

### Hybrid Aggregation

When using hybrid strategy:

1. Uses deterministic model as authoritative base
2. Merges additional signals from enhancement models
3. Weights base (70%) and enhancement (30%) confidence
4. Preserves deterministic classification
5. Enriches remediation with ML insights

## Performance Constraints

Configurable performance limits:

- `max_total_latency_ms`: 10000 (abort if exceeded)
- `max_cost_per_inference`: $0.10 (for future API-based models)
- `timeout_ms`: 30000 (per-model timeout)

## Multi-Model Report Structure

A multi-model inference result includes:

- `routing_decision`: Which strategy and models were selected, and why
- `model_results`: Individual results from each model
- `aggregated_result`: Final combined result
- `consensus_achieved`: Whether consensus threshold was met (if applicable)
- `confidence`: Final aggregated confidence score
- `execution_time_ms`: Total execution time
- `summary`: Human-readable summary

## Integration with Existing Workflow

Multi-model routing is an optional enhancement:

### Backward Compatibility

- Default behavior remains single deterministic model
- Existing `governance-sim` command unchanged
- New `governance-multi-model` command is opt-in

### Composability

Multi-model can be combined with:

- Threat detection: Include threat-detection model in consensus
- Drift analysis: Run after multi-model aggregation
- Remediation: Merge remediation from all models

## Adding New Models

To register a custom model:

```python
from assurancectl.model_registry import register_model, ModelMetadata

class CustomModel:
    def infer(self, config, proposal):
        # Implement inference logic
        return {
            "proposal_class": "high_impact",
            "risk_score": 35,
            "confidence": 0.82,
            # ...
        }

    def get_metadata(self):
        return ModelMetadata(
            model_id="custom-ml",
            model_version="1.0.0",
            model_type="ml",
            capabilities=["proposal-classification"],
            confidence_range=(0.6, 0.95),
            average_latency_ms=150.0,
            supported_proposal_classes=["standard", "high_impact", "safety_critical"],
            description="Custom ML-based classifier"
        )

    def health_check(self):
        return True  # Check if model is available

register_model("custom-ml", CustomModel())
```

## Future Enhancements

Planned improvements:

1. LLM integration for natural language summarization
2. External API-based models (OpenAI, Anthropic, etc.)
3. Async/parallel model execution for improved latency
4. Cost tracking and budget enforcement
5. A/B testing and model performance comparison
6. Automatic model selection based on historical accuracy
