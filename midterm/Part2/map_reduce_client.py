import asyncio
import json
import random
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
        # map_tasks = []
        # for i, chunk_url in enumerate(chunk_urls):
        #     mapper_ip = mapper_ips[i % len(mapper_ips)]

        #     async def make_request(idx=i, url=chunk_url, ip=mapper_ip):
        #         if idx == 1:
        #             await asyncio.sleep(10.0)
        #         return await self.http_get_json(
        #             f"http://{ip}:8080/map",
        #             params={"url": url}
        #         )

        #     map_tasks.append(make_request())
        
        # map_results = await asyncio.gather(*map_tasks)

        # for i, result in enumerate(map_results):
        #     if isinstance(result, dict) and "error" in result:
        #         alternate_mapper_ip = self._get_alternate_mapper(i)
        #         if alternate_mapper_ip:
        #             chunk_url = chunk_urls[i]
        #             alternative_result = await self.http_get_json(
        #                 f"http://{alternate_mapper_ip}:8080/map",
        #                 params={"url": chunk_url}
        #             )
        #             map_results[i] = alternative_result
        #         else:
        #             raise Exception(f"Map operation failed and no alternate mapper available")

        map_results = await self.process_map_chunks(chunk_urls, mapper_ips)

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
    
    def _get_alternate_mapper(self, failed_mapper_index: int) -> Optional[str]:
        available_mappers = [ip for i, ip in enumerate(self.mapper_ips) if i != failed_mapper_index]
        return available_mappers[random.randint(0, len(available_mappers)-1)] if available_mappers else None
    
    async def process_map_chunks(self, chunk_urls: List[str], mapper_ips: List[str]) -> List[Dict[str, Any]]:
        """Process map chunks with fault tolerance"""

        final_results = [None] * len(chunk_urls)
        completed_chunks = set()

        # Create initial mapping tasks
        active_tasks = {}
        for i, chunk_url in enumerate(chunk_urls):
            mapper_ip = mapper_ips[i % len(mapper_ips)]

            async def make_request():
                if i == 0:
                    await asyncio.sleep(10.0)
                return await self.http_get_json(
                    f"http://{mapper_ip}:8080/map",
                    params={"url": chunk_url}
                )

            task = asyncio.create_task(make_request())
            active_tasks[i] = task

        while len(completed_chunks) < len(chunk_urls):
            if not active_tasks:
                break  # No more active tasks to process
            done_tasks, _ = await asyncio.wait(active_tasks.values(), return_when=asyncio.FIRST_COMPLETED)
            
            for completed_task in done_tasks:
                # Find which chunk this task corresponds to
                chunk_idx = next((idx for idx, task in active_tasks.items() if task == completed_task), None)
                if chunk_idx is None:
                    continue

                try:
                    result = await completed_task
                    if isinstance(result, dict) and "error" in result:
                        alternate_mapper_ip = self._get_alternate_mapper(chunk_idx)
                        if alternate_mapper_ip:
                            chunk_url = chunk_urls[chunk_idx]
                            retry_task = asyncio.create_task(
                                self.http_get_json(
                                    f"http://{alternate_mapper_ip}:8080/map",
                                    params={"url": chunk_url}
                                )
                            )
                            active_tasks[chunk_idx] = retry_task
                            continue
                        else:
                            raise Exception(f"Map operation failed and no alternate mapper available")
                    else:
                        final_results[chunk_idx] = result
                    
                    completed_chunks.add(chunk_idx)
                    del active_tasks[chunk_idx]
                
                except Exception as e:
                    print(f"Error processing chunk {chunk_idx}: {e}")
                    final_results[chunk_idx] = {"error": str(e)}
                    completed_chunks.add(chunk_idx)
                    if chunk_idx in active_tasks:
                        del active_tasks[chunk_idx]

        return final_results

    async def http_get_json(self, url: str, params: Optional[Dict[str, str]] = None) -> Any:
        if params:
            url += "?" + urllib.parse.urlencode(params)
        
        print(f"HTTP GET: {url}")
        
        loop = asyncio.get_event_loop()
        
        def blocking_request():
            import urllib.error
            try:
                request = urllib.request.Request(url)
                with urllib.request.urlopen(request, timeout=300) as response:
                    if response.status != 200:
                        return {"error": f"HTTP request failed with status {response.status}"}
                    return json.load(response)
            except urllib.error.HTTPError as e:
                # Handle HTTP errors (4xx, 5xx)
                return {"error": f"HTTP {e.code} error: {e.reason}"}
            except Exception as e:
                return {"error": f"Request failed: {str(e)}"}
        
        try:
            return await loop.run_in_executor(None, blocking_request)
        except Exception as e:
            print(f"HTTP request failed for {url}: {e}")
            return {"error": f"Request execution failed: {str(e)}"}
    
    def print_result(self, result: MapReduceResult):
        print(f"MapReduce Result: {result.message}")
        print(f"Total Words: {result.total_words}")
        print(f"Unique Words: {result.unique_words}") 

if __name__ == "__main__":
    file_url = "s3://mapreduce-experiment-975050147762/50MB-TXT-FILE.txt"
    chunks = 6

    loop = asyncio.get_event_loop()

    task_manager = ECSTaskManager(s3_bucket="mapreduce-experiment-975050147762")

    # Clean S3 results before starting
    print("Cleaning previous S3 results...")
    loop.run_until_complete(task_manager.clean_s3_results())

    task_ips = loop.run_until_complete(task_manager.get_mapreduce_ips())

    mapper_ips = task_ips.mapper_ips[:chunks-1] + task_ips.mapfunctional_mapper_ips if task_ips.mapfunctional_mapper_ips else []

    client = MapReduceClient(
        splitter_ip=task_ips.splitter_ip,
        mapper_ips=mapper_ips,
        reducer_ip=task_ips.reducer_ip
    )
    result = loop.run_until_complete(client.perform_mapreduce(file_url, chunks=chunks))
    client.print_result(result)