import requests
import pytest

base_url = "https://dog.ceo/api"

def random_img():
    response = requests.get(f"{base_url}/breeds/image/random")
    data = response.json()
    return data.get('message')

def rand_img_code():
    response = requests.get(f"{base_url}/breeds/image/random")
    assert response.status_code == 200

def all_img_code():
    response = requests.get(f"{base_url}/breeds/list/all")
    assert response.status_code == 200

def test_url(random_dog_img):
    assert random_dog_img is not None
    assert random_dog_img.startswith("https://images.dog.ceo")

def test_format():
    response = requests.get(f"{base_url}/breeds/image/random")
    data = response.json()

    assert "message" in data
    img_url = data["message"]

    assert img_url.endswith(".jpg")

def get_specified_breed(breed):
    response = requests.get(f"{base_url}/breed/{breed}/images")
    data = response.json()

    assert "message" in data
    assert len(data["message"]) > 0

random_img()
rand_img_code()
all_img_code()
test_url("https://images.dog.ceo")
test_format()
get_specified_breed("terrier-russell")