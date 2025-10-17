# Crashing and Recovering Mappers in MapReduce
In this part of the midterm, I enhanced my MapReduce implementation to handle mapper failures gracefully. I created a "malfunctional" mapper that return 500 error during execution to simulate real-world failures. 

For an example file, I count it on a local script, showing results:  
<img width="686" height="96" alt="image" src="https://github.com/user-attachments/assets/7cbbe5a7-f34e-43b0-8978-94cefe0a11e7" />  
However if I replace 1 mapper with a malfunctional-mapper, 
<img width="998" height="444" alt="image" src="https://github.com/user-attachments/assets/a6aeb597-9366-4514-ad39-7ef64b889969" />
it returns:  
<img width="689" height="126" alt="image" src="https://github.com/user-attachments/assets/df9fdc77-0980-4320-a294-5a1ef760f56b" />  
It indicates that if a mapper failed, the task assigned to it will not be done correctly, so the total result will lose part of result.

The failure comes from:
```python
map_tasks = []
for i, chunk_url in enumerate(chunk_urls):
    mapper_ip = mapper_ips[i % len(mapper_ips)]
    map_tasks.append(self.http_get_json(
        f"http://{mapper_ip}:8080/map",
        params={"url": chunk_url}
    ))

map_results = await asyncio.gather(*map_tasks)
```
`map_results` contains 5 valid result.  
<img width="444" height="165" alt="image" src="https://github.com/user-attachments/assets/c6d4d56a-ae97-47d5-a02e-cddf3bc2db6b" />  
So the reducer would only fetch 5 intermediate results.

### Fix:
After gathering map results, try to find the error type's result, then assign the map task to another available mapper, it can address current problem.
```python
for i, result in enumerate(map_results):
    if isinstance(result, dict) and "error" in result:
        alternate_mapper_ip = self._get_alternate_mapper(i)
        if alternate_mapper_ip:
            chunk_url = chunk_urls[i]
            alternative_result = await self.http_get_json(
                f"http://{alternate_mapper_ip}:8080/map",
                params={"url": chunk_url}
            )
            map_results[i] = alternative_result
        else:
            raise Exception(f"Map operation failed and no alternate mapper available")
```
<img width="692" height="79" alt="image" src="https://github.com/user-attachments/assets/9d037ecd-61c6-4f40-8f1b-6d33c16beb9f" />



