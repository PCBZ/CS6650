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
            # Simplified for maximum RPS - minimal processing
            if response.status_code == 200:
                response.success()
            else:
                response.failure(f"Status: {response.status_code}")