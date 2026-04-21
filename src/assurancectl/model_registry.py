"""Model registry for managing multiple inference engines."""

from __future__ import annotations

import time
from dataclasses import dataclass, field
from datetime import datetime, timezone
from typing import Any, Callable, Protocol


class InferenceModel(Protocol):
    """Protocol for inference models."""

    def infer(self, config: Any, proposal: dict[str, Any]) -> dict[str, Any]:
        """Run inference on a proposal."""
        ...

    def get_metadata(self) -> ModelMetadata:
        """Get model metadata."""
        ...

    def health_check(self) -> bool:
        """Check if model is healthy and available."""
        ...


@dataclass
class ModelMetadata:
    """Metadata about an inference model."""

    model_id: str
    model_version: str
    model_type: str
    capabilities: list[str]
    confidence_range: tuple[float, float]
    average_latency_ms: float
    supported_proposal_classes: list[str]
    description: str


@dataclass
class ModelHealthStatus:
    """Health status tracking for a model."""

    model_id: str
    is_available: bool
    last_check: datetime
    failure_count: int
    last_failure: datetime | None = None
    last_success: datetime | None = None


@dataclass
class ModelRegistry:
    """Registry for managing multiple inference models."""

    models: dict[str, InferenceModel] = field(default_factory=dict)
    health_status: dict[str, ModelHealthStatus] = field(default_factory=dict)
    _health_check_enabled: bool = True
    _failure_threshold: int = 3
    _recovery_cooldown_seconds: int = 300

    def register_model(
        self,
        model_id: str,
        model: InferenceModel,
    ) -> None:
        """Register a new model in the registry."""
        self.models[model_id] = model
        self.health_status[model_id] = ModelHealthStatus(
            model_id=model_id,
            is_available=True,
            last_check=datetime.now(UTC),
            failure_count=0,
        )

    def get_model(self, model_id: str) -> InferenceModel | None:
        """Get a model by ID if it's available."""
        if model_id not in self.models:
            return None

        if not self._health_check_enabled:
            return self.models[model_id]

        status = self.health_status.get(model_id)
        if status and not status.is_available:
            # Check if cooldown period has passed
            if status.last_failure:
                elapsed = (datetime.now(UTC) - status.last_failure).total_seconds()
                if elapsed < self._recovery_cooldown_seconds:
                    return None

        return self.models[model_id]

    def check_model_health(self, model_id: str) -> bool:
        """Check health of a specific model."""
        if model_id not in self.models:
            return False

        model = self.models[model_id]
        status = self.health_status[model_id]
        now = datetime.now(UTC)

        try:
            is_healthy = model.health_check()
            status.last_check = now

            if is_healthy:
                status.is_available = True
                status.failure_count = 0
                status.last_success = now
                return True
            else:
                status.failure_count += 1
                if status.failure_count >= self._failure_threshold:
                    status.is_available = False
                    status.last_failure = now
                return False

        except Exception:
            status.failure_count += 1
            status.last_check = now
            if status.failure_count >= self._failure_threshold:
                status.is_available = False
                status.last_failure = now
            return False

    def get_available_models(self) -> list[str]:
        """Get list of currently available model IDs."""
        available = []
        for model_id in self.models:
            if self.get_model(model_id) is not None:
                available.append(model_id)
        return available

    def get_models_by_capability(self, capability: str) -> list[str]:
        """Get models that support a specific capability."""
        matching = []
        for model_id, model in self.models.items():
            metadata = model.get_metadata()
            if capability in metadata.capabilities:
                if self.get_model(model_id) is not None:
                    matching.append(model_id)
        return matching

    def configure_health_checks(
        self,
        *,
        enabled: bool = True,
        failure_threshold: int = 3,
        recovery_cooldown_seconds: int = 300,
    ) -> None:
        """Configure health check settings."""
        self._health_check_enabled = enabled
        self._failure_threshold = failure_threshold
        self._recovery_cooldown_seconds = recovery_cooldown_seconds


# Default registry instance
_default_registry: ModelRegistry | None = None


def get_default_registry() -> ModelRegistry:
    """Get the default global model registry."""
    global _default_registry
    if _default_registry is None:
        _default_registry = ModelRegistry()
    return _default_registry


def register_model(model_id: str, model: InferenceModel) -> None:
    """Register a model in the default registry."""
    registry = get_default_registry()
    registry.register_model(model_id, model)
