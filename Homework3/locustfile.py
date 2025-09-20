from locust import HttpUser, task, between, FastHttpUser

class AlbumTestUser(HttpUser):
    wait_time = between(1, 3) # Simulate user wait time between tasks

    def on_start(self):
        """ Called when a simulated user starts """
        print("Starting a new user session")
    
    @task(3)
    def get_albums(self):
        """ Task to get albums """
        with self.client.get("/albums", catch_response=True) as response:
            if response.status_code == 200:
                response.success()
            else:
                response.failure(f"Failed with status code {response.status_code}")

    @task(1)
    def post_album(self):
        """ Task to post a new album """
        post_data = {
            "id": "4",
            "title": "The Modern Sound of Betty Carter",
            "artist": "Betty Carter",
            "price": 49.99
        }
        headers = {
            "Content-Type": "application/json"
        }
        with self.client.post("/albums", json=post_data, headers=headers, catch_response=True) as response:
            if response.status_code in [200, 201]:
                response.success()
            else:
                response.failure(f"Failed with status code {response.status_code}")