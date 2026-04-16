"""Unit tests for ai-service endpoints."""

from fastapi.testclient import TestClient
import pytest

from main import app
from routers import description_generator, image_generator


@pytest.fixture
def client() -> TestClient:
    """Provide a FastAPI test client."""
    return TestClient(app)


@pytest.fixture(autouse=True)
def clear_relevant_env(monkeypatch: pytest.MonkeyPatch) -> None:
    """Clear environment variables used by endpoint routing logic."""
    for variable in [
        "AZURE_OPENAI_DALLE_ENDPOINT",
        "AZURE_OPENAI_ENDPOINT",
        "AZURE_OPENAI_DALLE_DEPLOYMENT_NAME",
        "USE_AZURE_OPENAI",
        "USE_LOCAL_LLM",
        "USE_AZURE_AD",
    ]:
        monkeypatch.delenv(variable, raising=False)


def test_health_reports_description_capability_by_default(client: TestClient) -> None:
    response = client.get("/health")

    assert response.status_code == 200
    assert response.json()["capabilities"] == ["description"]


def test_health_reports_image_capability_when_env_set(
    client: TestClient, monkeypatch: pytest.MonkeyPatch
) -> None:
    monkeypatch.setenv("AZURE_OPENAI_ENDPOINT", "https://example.openai.azure.com")
    monkeypatch.setenv("AZURE_OPENAI_DALLE_DEPLOYMENT_NAME", "dalle-deployment")

    response = client.get("/health")

    assert response.status_code == 200
    assert response.json()["capabilities"] == ["description", "image"]


def test_generate_description_uses_mocked_openai_handler(
    client: TestClient, monkeypatch: pytest.MonkeyPatch
) -> None:
    expected_description = "Mocked product description"

    def mock_handle_openai(user_prompt: str) -> str:
        assert "dog toy" in user_prompt
        return expected_description

    monkeypatch.setattr(description_generator, "_handle_openai", mock_handle_openai)

    response = client.post(
        "/generate/description",
        json={"name": "dog toy", "tags": ["durable", "chewable"]},
    )

    assert response.status_code == 200
    assert response.json() == {"description": expected_description}


def test_generate_image_uses_mocked_azure_handler(
    client: TestClient, monkeypatch: pytest.MonkeyPatch
) -> None:
    expected_url = "https://images.example/mock.png"

    def mock_handle_azure_openai(user_prompt: str, use_azure_ad: bool) -> str:
        assert "cat bed" in user_prompt
        assert use_azure_ad is False
        return expected_url

    monkeypatch.setattr(image_generator, "_handle_azure_openai", mock_handle_azure_openai)

    response = client.post(
        "/generate/image",
        json={"name": "cat bed", "description": "plush and cozy"},
    )

    assert response.status_code == 200
    assert response.json() == {"image": expected_url}


def test_generate_image_returns_500_when_handler_fails(
    client: TestClient, monkeypatch: pytest.MonkeyPatch
) -> None:
    def mock_handle_azure_openai(user_prompt: str, use_azure_ad: bool) -> str:
        raise RuntimeError("external API unavailable")

    monkeypatch.setattr(image_generator, "_handle_azure_openai", mock_handle_azure_openai)

    response = client.post(
        "/generate/image",
        json={"name": "bird perch", "description": "wooden stand"},
    )

    assert response.status_code == 500
    assert "Error generating image" in response.json()["detail"]
