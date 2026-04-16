from fastapi.testclient import TestClient
import main
from routers import description_generator, image_generator


client = TestClient(main.app)


def test_health_endpoint_responds() -> None:
    response = client.get("/health")
    assert response.status_code == 200
    assert response.json()["status"] == "ok"


def test_generate_description_endpoint_responds(monkeypatch) -> None:
    monkeypatch.setattr(description_generator, "_handle_openai", lambda _prompt: "test description")
    response = client.post("/generate/description", json={"name": "Dog Toy", "tags": ["durable", "chew"]})
    assert response.status_code == 200
    assert response.json() == {"description": "test description"}


def test_generate_image_endpoint_responds(monkeypatch) -> None:
    monkeypatch.setattr(image_generator, "_handle_azure_openai", lambda _prompt, _use_azure_ad: "https://example.com/image.png")
    response = client.post("/generate/image", json={"name": "Dog Toy", "description": "A toy for dogs"})
    assert response.status_code == 200
    assert response.json() == {"image": "https://example.com/image.png"}
