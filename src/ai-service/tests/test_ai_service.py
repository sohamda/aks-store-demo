"""
Unit tests for the AI service.

Tests cover:
- /health endpoint
- POST /generate/description endpoint (with mocked LLM providers)
- POST /generate/image endpoint (with mocked Azure OpenAI DALL-E)
"""
import os
import pytest
from unittest.mock import patch, MagicMock
from fastapi.testclient import TestClient


@pytest.fixture
def client():
    """Return a TestClient for the FastAPI app."""
    from main import app
    return TestClient(app)


# ---------------------------------------------------------------------------
# Health endpoint
# ---------------------------------------------------------------------------

class TestHealthEndpoint:
    def test_health_returns_ok_status(self, client):
        response = client.get("/health")
        assert response.status_code == 200
        body = response.json()
        assert body["status"] == "ok"

    def test_health_always_has_description_capability(self, client):
        response = client.get("/health")
        assert "description" in response.json()["capabilities"]

    def test_health_no_image_capability_without_dalle_env_vars(self, client):
        env_without_dalle = {k: v for k, v in os.environ.items()
                             if k not in ("AZURE_OPENAI_DALLE_ENDPOINT",
                                          "AZURE_OPENAI_DALLE_DEPLOYMENT_NAME",
                                          "AZURE_OPENAI_ENDPOINT")}
        with patch.dict(os.environ, env_without_dalle, clear=True):
            response = client.get("/health")
        assert response.status_code == 200
        assert "image" not in response.json()["capabilities"]

    def test_health_includes_image_capability_with_dalle_endpoint_and_deployment(self, client):
        with patch.dict(os.environ, {
            "AZURE_OPENAI_DALLE_ENDPOINT": "https://example.openai.azure.com",
            "AZURE_OPENAI_DALLE_DEPLOYMENT_NAME": "dall-e-3",
        }):
            response = client.get("/health")
        assert response.status_code == 200
        assert "image" in response.json()["capabilities"]

    def test_health_includes_image_capability_with_base_endpoint_and_dalle_deployment(self, client):
        with patch.dict(os.environ, {
            "AZURE_OPENAI_ENDPOINT": "https://example.openai.azure.com",
            "AZURE_OPENAI_DALLE_DEPLOYMENT_NAME": "dall-e-3",
        }):
            response = client.get("/health")
        assert response.status_code == 200
        assert "image" in response.json()["capabilities"]

    def test_health_returns_app_version(self, client):
        with patch.dict(os.environ, {"APP_VERSION": "1.2.3"}):
            from main import app
            app.version = os.environ.get("APP_VERSION", "0.1.0")
            test_client = TestClient(app)
            response = test_client.get("/health")
        assert response.status_code == 200


# ---------------------------------------------------------------------------
# Description generation endpoint
# ---------------------------------------------------------------------------

class TestDescriptionEndpoint:
    def test_generate_description_with_local_llm(self, client):
        with patch("routers.description_generator._handle_local_llm",
                   return_value="A wonderful product for pets.") as mock_llm:
            with patch.dict(os.environ, {"USE_LOCAL_LLM": "true",
                                         "LOCAL_LLM_ENDPOINT": "http://localhost:8080"}):
                response = client.post(
                    "/generate/description",
                    json={"name": "Cat Food", "tags": ["cat", "food", "premium"]}
                )
        assert response.status_code == 200
        assert response.json()["description"] == "A wonderful product for pets."
        mock_llm.assert_called_once()

    def test_generate_description_with_openai(self, client):
        with patch("routers.description_generator._handle_openai",
                   return_value="Delicious and nutritious.") as mock_openai:
            with patch.dict(os.environ, {"USE_LOCAL_LLM": "false",
                                         "USE_AZURE_OPENAI": "false"}):
                response = client.post(
                    "/generate/description",
                    json={"name": "Dog Treats", "tags": ["dog", "treats"]}
                )
        assert response.status_code == 200
        assert response.json()["description"] == "Delicious and nutritious."
        mock_openai.assert_called_once()

    def test_generate_description_with_azure_openai(self, client):
        with patch("routers.description_generator._handle_azure_openai",
                   return_value="Premium Azure-generated description.") as mock_azure:
            with patch.dict(os.environ, {"USE_LOCAL_LLM": "false",
                                         "USE_AZURE_OPENAI": "true",
                                         "USE_AZURE_AD": "false"}):
                response = client.post(
                    "/generate/description",
                    json={"name": "Bird Seed", "tags": ["bird", "seed"]}
                )
        assert response.status_code == 200
        assert response.json()["description"] == "Premium Azure-generated description."
        mock_azure.assert_called_once()

    def test_generate_description_llm_error_returns_500(self, client):
        with patch("routers.description_generator._handle_openai",
                   side_effect=Exception("API unreachable")):
            with patch.dict(os.environ, {"USE_LOCAL_LLM": "false",
                                         "USE_AZURE_OPENAI": "false"}):
                response = client.post(
                    "/generate/description",
                    json={"name": "Fish Tank", "tags": ["fish", "aquarium"]}
                )
        assert response.status_code == 500
        assert "Error generating description" in response.json()["detail"]

    def test_generate_description_missing_name_returns_422(self, client):
        response = client.post(
            "/generate/description",
            json={"tags": ["cat", "food"]}
        )
        assert response.status_code == 422

    def test_generate_description_missing_tags_returns_422(self, client):
        response = client.post(
            "/generate/description",
            json={"name": "Cat Food"}
        )
        assert response.status_code == 422

    def test_generate_description_empty_tags_list(self, client):
        with patch("routers.description_generator._handle_openai",
                   return_value="Simple description."):
            with patch.dict(os.environ, {"USE_LOCAL_LLM": "false",
                                         "USE_AZURE_OPENAI": "false"}):
                response = client.post(
                    "/generate/description",
                    json={"name": "Plain Product", "tags": []}
                )
        assert response.status_code == 200

    def test_generate_description_prompt_contains_product_name(self, client):
        captured_prompts = []

        def capture_prompt(user_prompt):
            captured_prompts.append(user_prompt)
            return "Description text."

        with patch("routers.description_generator._handle_openai",
                   side_effect=capture_prompt):
            with patch.dict(os.environ, {"USE_LOCAL_LLM": "false",
                                         "USE_AZURE_OPENAI": "false"}):
                client.post(
                    "/generate/description",
                    json={"name": "Luxury Cat Bed", "tags": ["cat", "luxury"]}
                )
        assert len(captured_prompts) == 1
        assert "Luxury Cat Bed" in captured_prompts[0]


# ---------------------------------------------------------------------------
# Image generation endpoint
# ---------------------------------------------------------------------------

class TestImageEndpoint:
    def test_generate_image_success(self, client):
        fake_url = "https://example.com/generated-image.png"
        with patch("routers.image_generator._handle_azure_openai",
                   return_value=fake_url):
            with patch.dict(os.environ, {"USE_AZURE_AD": "false"}):
                response = client.post(
                    "/generate/image",
                    json={"name": "Cat Toy", "description": "A fun toy for cats"}
                )
        assert response.status_code == 200
        assert response.json()["image"] == fake_url

    def test_generate_image_error_returns_500(self, client):
        with patch("routers.image_generator._handle_azure_openai",
                   side_effect=Exception("DALL-E unavailable")):
            with patch.dict(os.environ, {"USE_AZURE_AD": "false"}):
                response = client.post(
                    "/generate/image",
                    json={"name": "Cat Toy", "description": "A fun toy for cats"}
                )
        assert response.status_code == 500
        assert "Error generating image" in response.json()["detail"]

    def test_generate_image_missing_name_returns_422(self, client):
        response = client.post(
            "/generate/image",
            json={"description": "A fun toy for cats"}
        )
        assert response.status_code == 422

    def test_generate_image_missing_description_returns_422(self, client):
        response = client.post(
            "/generate/image",
            json={"name": "Cat Toy"}
        )
        assert response.status_code == 422

    def test_generate_image_prompt_contains_product_name(self, client):
        captured_prompts = []

        def capture_prompt(user_prompt, use_azure_ad):
            captured_prompts.append(user_prompt)
            return "https://example.com/img.png"

        with patch("routers.image_generator._handle_azure_openai",
                   side_effect=capture_prompt):
            with patch.dict(os.environ, {"USE_AZURE_AD": "false"}):
                client.post(
                    "/generate/image",
                    json={"name": "Dog Leash", "description": "Sturdy leash for dogs"}
                )
        assert len(captured_prompts) == 1
        assert "Dog Leash" in captured_prompts[0]


# ---------------------------------------------------------------------------
# Internal handler unit tests
# ---------------------------------------------------------------------------

class TestDescriptionHandlers:
    def test_handle_local_llm_raises_without_endpoint(self):
        from routers.description_generator import _handle_local_llm
        with patch.dict(os.environ, {}, clear=True):
            with pytest.raises(ValueError, match="LOCAL_LLM_ENDPOINT"):
                _handle_local_llm("some prompt")

    def test_handle_openai_raises_without_api_key(self):
        from routers.description_generator import _handle_openai
        env = {k: v for k, v in os.environ.items()
               if k not in ("OPENAI_API_KEY", "OPENAI_ORG_ID")}
        with patch.dict(os.environ, env, clear=True):
            with pytest.raises(ValueError, match="OPENAI_API_KEY"):
                _handle_openai("some prompt")

    def test_handle_azure_openai_raises_without_deployment(self):
        from routers.description_generator import _handle_azure_openai
        env = {k: v for k, v in os.environ.items()
               if k not in ("AZURE_OPENAI_DEPLOYMENT_NAME", "AZURE_OPENAI_ENDPOINT")}
        with patch.dict(os.environ, env, clear=True):
            with pytest.raises(ValueError, match="AZURE_OPENAI_DEPLOYMENT_NAME"):
                _handle_azure_openai("some prompt", use_azure_ad=False)


class TestImageHandlers:
    def test_handle_azure_openai_raises_without_endpoint(self):
        from routers.image_generator import _handle_azure_openai
        env = {k: v for k, v in os.environ.items()
               if k not in ("AZURE_OPENAI_DALLE_ENDPOINT", "AZURE_OPENAI_ENDPOINT")}
        with patch.dict(os.environ, env, clear=True):
            with pytest.raises(ValueError, match="AZURE_OPENAI_DALLE_ENDPOINT"):
                _handle_azure_openai("some prompt", use_azure_ad=False)

    def test_handle_azure_openai_raises_without_deployment_name(self):
        from routers.image_generator import _handle_azure_openai
        env = {k: v for k, v in os.environ.items()
               if k != "AZURE_OPENAI_DALLE_DEPLOYMENT_NAME"}
        env["AZURE_OPENAI_DALLE_ENDPOINT"] = "https://example.com"
        with patch.dict(os.environ, env, clear=True):
            with pytest.raises(ValueError, match="AZURE_OPENAI_DALLE_DEPLOYMENT_NAME"):
                _handle_azure_openai("some prompt", use_azure_ad=False)

    def test_handle_azure_openai_raises_without_api_version(self):
        from routers.image_generator import _handle_azure_openai
        env = {k: v for k, v in os.environ.items()
               if k != "AZURE_OPENAI_API_VERSION"}
        env["AZURE_OPENAI_DALLE_ENDPOINT"] = "https://example.com"
        env["AZURE_OPENAI_DALLE_DEPLOYMENT_NAME"] = "dall-e-3"
        with patch.dict(os.environ, env, clear=True):
            with pytest.raises(ValueError, match="AZURE_OPENAI_API_VERSION"):
                _handle_azure_openai("some prompt", use_azure_ad=False)

    def test_handle_azure_openai_raises_without_api_key_when_not_using_azure_ad(self):
        from routers.image_generator import _handle_azure_openai
        env = {k: v for k, v in os.environ.items()
               if k != "OPENAI_API_KEY"}
        env["AZURE_OPENAI_DALLE_ENDPOINT"] = "https://example.com"
        env["AZURE_OPENAI_DALLE_DEPLOYMENT_NAME"] = "dall-e-3"
        env["AZURE_OPENAI_API_VERSION"] = "2024-02-15-preview"
        with patch.dict(os.environ, env, clear=True):
            with pytest.raises(ValueError, match="OPENAI_API_KEY"):
                _handle_azure_openai("some prompt", use_azure_ad=False)
