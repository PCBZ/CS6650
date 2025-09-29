#!/usr/bin/env python3

import asyncio
import time
import matplotlib.pyplot as plt
from typing import List, Dict
from fetch_ecs_task import ECSTaskManager
from map_reduce_client import MapReduceClient

class MapReducePerformanceTester:
    
    async def test_different_mappers(self, input_url: str, mapper_counts: List[int]) -> Dict[int, float]:
        """Test MapReduce with different mapper counts, return duration results"""
        
        # Get task IPs
        task_manager = ECSTaskManager(s3_bucket="mapreduce-experiment-975050147762")
        task_ips = await task_manager.get_mapreduce_ips()
        
        results = {}

        verified_word_counts = None
        
        for mapper_count in mapper_counts:
            print(f"\nTesting with {mapper_count} mappers...")

            if mapper_count > len(task_ips.mapper_ips):
                print(f"  Not enough mappers available ({len(task_ips.mapper_ips)}). Skipping.")
                continue
            
            # Use limited number of mappers
            limited_mappers = task_ips.mapper_ips[:mapper_count]

            await task_manager.clean_s3_results()
            
            client = MapReduceClient(
                splitter_ip=task_ips.splitter_ip,
                mapper_ips=limited_mappers,
                reducer_ip=task_ips.reducer_ip
            )
            
            # Measure duration
            start_time = time.time()
            map_reduce_result = await client.perform_mapreduce(input_url, mapper_count)
            
            word_counts = map_reduce_result.total_words

            if verified_word_counts is None:
                verified_word_counts = word_counts
            elif verified_word_counts != word_counts:
                print("  Warning: Word counts differ from previous runs!")

            duration = time.time() - start_time
            
            results[mapper_count] = duration
            print(f"Duration: {duration:.2f}s")
            
            # Wait between tests
            if mapper_count != mapper_counts[-1]:
                await asyncio.sleep(5)
        
        return results
    
    def plot_duration_curve(self, results: Dict[int, float]):
        """Create duration vs mapper count curve"""
        
        mapper_counts = sorted(results.keys())
        durations = [results[count] for count in mapper_counts]
        
        plt.figure(figsize=(10, 6))
        plt.plot(mapper_counts, durations, 'b-o', linewidth=2, markersize=8)
        plt.xlabel('Number of Mappers')
        plt.ylabel('Duration (seconds)')
        plt.title('MapReduce Duration vs Mapper Count (164K)')
        plt.grid(True, alpha=0.3)
        
        plt.show()
    

async def main():
    # file_url= "https://www.gutenberg.org/files/100/100-0.txt"
    # file_url = "s3://mapreduce-experiment-975050147762/20MB-file.txt"
    # file_url = "s3://mapreduce-experiment-975050147762/50MB-TXT-FILE.txt"
    # file_url = "https://mapreduce-experiment-975050147762.s3.us-west-2.amazonaws.com/50MB-TXT-FILE.txt"
    # file_url = "https://drive.google.com/file/d/1XvQpU2U4QDBGt__uwsksI3gbpEOns-PX/view?usp=drive_lin"
    
    # file_url = "https://www.sampledocs.in/DownloadFiles/SampleFile?filename=sample_file_50mb_sampledocs&ext=txt"
    file_url = "https://raw.githubusercontent.com/teropa/nlp/master/resources/corpora/gutenberg/shakespeare-hamlet.txt"
    mapper_counts = [1, 2, 3, 4, 5, 6]
    
    tester = MapReducePerformanceTester()
    
    # Run tests
    results = await tester.test_different_mappers(file_url, mapper_counts)
    
    # Show results
    tester.plot_duration_curve(results)

if __name__ == "__main__":
    asyncio.run(main())