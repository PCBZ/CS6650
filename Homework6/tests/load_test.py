"""
Product API Search Load Testing Suite
CS6650 Distributed Systems - Search Performance Analysis

Focused testing for the search endpoint using FastHttpUser for optimal performance.
"""

from locust import FastHttpUser, task, between
import random
import json
import time


class SearchAPIUser(FastHttpUser):
    """
    FastHttpUser focused on search API testing
    Uses httpx under the hood for better async performance
    """
    wait_time = between(1, 3)
    
    @task
    def search_products(self):
        """GET /v1/products/search?q={query} - Test the search endpoint"""
        search_queries = [
            "Alpha",        # Brand search
            "Beta",         # Brand search  
            "Electronics",  # Category search
            "Books",        # Category search
            "Home",         # Category search
            "Sports",       # Category search
            "Product",      # General product search
            "Gamma",        # Brand search
            "Clothing",     # Category search
            "NonExistent",  # Search with no results
        ]
        
        query = random.choice(search_queries)
        start_time = time.time()
        
        with self.client.get(f"/v1/products/search?q={query}", catch_response=True) as response:
            response_time = (time.time() - start_time) * 1000
            
            if response.status_code == 200:
                try:
                    data = response.json()
                    # Validate response structure
                    if "products" in data and "total_found" in data and "search_time" in data:
                        # Check that we get max 20 results
                        if len(data["products"]) <= 20:
                            # Monitor search performance
                            if "search_time" in data:
                                server_search_time = float(data["search_time"].replace("ms", ""))
                                if server_search_time > 10:  # Alert if search takes > 10ms
                                    print(f"Slow search detected: {server_search_time}ms for query '{query}'")
                            
                            response.success()
                        else:
                            response.failure(f"Too many products returned: {len(data['products'])}")
                    else:
                        response.failure("Invalid response structure")
                except json.JSONDecodeError:
                    response.failure("Invalid JSON response")
            else:
                response.failure(f"Got status code {response.status_code}")