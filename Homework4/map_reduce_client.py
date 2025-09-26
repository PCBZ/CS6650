import asyncio
import json
import time
import urllib.request
import urllib.parse
from dataclasses import dataclass
from typing import List, Dict, Any, Optional

from fetch_ecs_task import ECSTaskManager

@dataclass
class ECSTask:
    task_arn: str
    task_type: str
    public_ip: str
    status: str

@dataclass
class MapReduceResult:
    words: Dict[str, int]
    total_words: int
    unique_words: int
    message: str

class MapReduceClient:

    def __init__(self, splitter_ip: str, mapper_ips: List[str], reducer_ip: str):
        self.splitter_ip = splitter_ip
        self.mapper_ips = mapper_ips
        self.reducer_ip = reducer_ip

    async def perform_mapreduce(self, input_url: str, chunks: int = 3) -> MapReduceResult:

        splitter_ip = self.splitter_ip
        mapper_ips = self.mapper_ips
        reducer_ip = self.reducer_ip

        start_time = time.time()

        ### Perform Map Reduce Steps ###
        # 1. Split the input file
        split_result = await self.http_get_json(
            f"http://{splitter_ip}:8080/split",
            params={"url": input_url, "chunks": str(chunks)}
        )

        chunk_urls = split_result.get("chunk_urls", [])

        # 2. Map each chunk
        map_tasks = []
        for i, chunk_url in enumerate(chunk_urls):
            mapper_ip = mapper_ips[i % len(mapper_ips)]
            map_tasks.append(self.http_get_json(
                f"http://{mapper_ip}:8080/map",
                params={"url": chunk_url}
            ))
        
        map_results = await asyncio.gather(*map_tasks)
        print(f"Completed {len(map_results)} map operations")

        # 3. Reduce the mapped results
        reduce_result = await self.http_get_json(
            f"http://{reducer_ip}:8080/reduce"
        )

        duration = time.time() - start_time

        print(f"MapReduce completed in {duration:.2f} seconds")

        return MapReduceResult(
            words=reduce_result.get("words", {}),
            total_words=reduce_result.get("total_words", 0),
            unique_words=reduce_result.get("unique_words", 0),
            message=reduce_result.get("message", "")
        )
    
    async def http_get_json(self, url: str, params: Optional[Dict[str, str]] = None) -> Any:
        if params:
            url += "?" + urllib.parse.urlencode(params)
        
        print(f"HTTP GET: {url}")
        
        loop = asyncio.get_event_loop()
        
        def blocking_request():
            with urllib.request.urlopen(url) as response:
                if response.status != 200:
                    raise Exception(f"HTTP request failed with status {response.status}")
                return json.load(response)
        
        return await loop.run_in_executor(None, blocking_request)
    
    def print_result(self, result: MapReduceResult):
        print(f"MapReduce Result: {result.message}")
        print(f"Total Words: {result.total_words}")
        print(f"Unique Words: {result.unique_words}") 

if __name__ == "__main__":
    file_url = "https://raw.githubusercontent.com/teropa/nlp/master/resources/corpora/gutenberg/shakespeare-hamlet.txt"
    chunks = 4

    loop = asyncio.get_event_loop()

    task_manager = ECSTaskManager()
    task_ips = loop.run_until_complete(task_manager.get_mapreduce_ips())


    client = MapReduceClient(
        splitter_ip=task_ips.splitter_ip,
        mapper_ips=task_ips.mapper_ips,
        reducer_ip=task_ips.reducer_ip
    )
    result = loop.run_until_complete(client.perform_mapreduce(file_url, chunks=chunks))
    client.print_result(result)