#!/usr/bin/env python3

import asyncio
import json
from dataclasses import dataclass
from typing import List, Dict, Any, Optional

@dataclass
class ECSTask:
    task_arn: str
    task_type: str
    public_ip: str
    status: str

@dataclass
class TaskIPs:
    splitter_ip: str
    mapper_ips: List[str]
    reducer_ip: str

class ECSTaskManager:

    def __init__(self, s3_bucket: str, cluster_name: str = "hello-cluster", region: str = "us-west-2"):
        self.s3_bucket = s3_bucket
        self.cluster_name = cluster_name
        self.region = region

    async def get_all_tasks(self) -> List[ECSTask]:
        """Get all running ECS tasks with their public IPs"""
        cmd = [
            "aws", "ecs", "list-tasks",
            "--cluster", self.cluster_name,
            "--region", self.region,
            "--query", "taskArns[]",
            "--output", "text"
        ]

        process = await asyncio.create_subprocess_exec(
            *cmd,
            stdout=asyncio.subprocess.PIPE,
            stderr=asyncio.subprocess.PIPE
        )
        stdout, stderr = await process.communicate()

        if process.returncode != 0:
            raise Exception(f"Error listing tasks: {stderr.decode().strip()}")

        # Parse task ARNs
        output = stdout.decode().strip()
        if not output:
            print("No tasks found in cluster")
            return []
        
        task_arns = []
        for line in output.split('\n'):
            line = line.strip()
            if line and line != 'None':
                arns = line.split('\t') if '\t' in line else [line]
                for arn in arns:
                    arn = arn.strip()
                    if arn and arn != 'None':
                        task_arns.append(arn)
        
        print(f"Found {len(task_arns)} task ARNs")
        
        if not task_arns:
            return []

        return await self._describe_tasks(task_arns)
    
    async def _describe_tasks(self, task_arns: List[str]) -> List[ECSTask]:
        """Get detailed information for each task"""
        if not task_arns:
            return []

        # Validate task ARNs
        valid_arns = []
        for arn in task_arns:
            if arn and ('arn:aws:ecs:' in arn or '/' in arn or len(arn) > 10):
                valid_arns.append(arn)
        
        if not valid_arns:
            return []

        cmd = [
            "aws", "ecs", "describe-tasks",
            "--cluster", self.cluster_name,
            "--region", self.region,
            "--tasks"
        ] + valid_arns + [
            "--query", "tasks[]",
            "--output", "json"
        ]

        process = await asyncio.create_subprocess_exec(
            *cmd,
            stdout=asyncio.subprocess.PIPE,
            stderr=asyncio.subprocess.PIPE
        )
        stdout, stderr = await process.communicate()

        if process.returncode != 0:
            print(f"Error describing tasks: {stderr.decode().strip()}")
            return []

        tasks_info = json.loads(stdout.decode())
        tasks = []

        for task in tasks_info:
            task_arn = task.get("taskArn", "")
            last_status = task.get("lastStatus", "")
            public_ip = await self._extract_public_ip(task)
            task_type = self._extract_task_type(task.get("taskDefinitionArn", ""))

            if last_status == "RUNNING" and public_ip:
                tasks.append(ECSTask(
                    task_arn=task_arn,
                    task_type=task_type,
                    public_ip=public_ip,
                    status=last_status
                ))

        return tasks

    def _extract_task_type(self, task_def_arn: str) -> str:
        """Extract task type from task definition ARN"""
        try:
            parts = task_def_arn.split("/")
            if len(parts) > 1:
                name_version = parts[-1]
                name = name_version.split(":")[0]
                if name.endswith("-task"):
                    name = name[:-5]
                return name
        except:
            pass
        return "unknown"

    async def _extract_public_ip(self, task_data: Dict) -> Optional[str]:
        """Extract public IP from task data"""
        attachments = task_data.get("attachments", [])
        for attachment in attachments:
            if attachment.get("type") == "ElasticNetworkInterface":
                details = attachment.get("details", [])
                for detail in details:
                    if detail.get("name") == "networkInterfaceId":
                        eni_id = detail.get("value")
                        return await self._get_public_ip_from_eni(eni_id)
        return None
    
    async def _get_public_ip_from_eni(self, eni_id: str) -> Optional[str]:
        """Get public IP from network interface ID"""
        cmd = [
            "aws", "ec2", "describe-network-interfaces",
            "--network-interface-ids", eni_id,
            "--region", self.region,
            "--query", "NetworkInterfaces[0].Association.PublicIp",
            "--output", "text"
        ]

        process = await asyncio.create_subprocess_exec(
            *cmd,
            stdout=asyncio.subprocess.PIPE,
            stderr=asyncio.subprocess.PIPE
        )
        stdout, stderr = await process.communicate()

        if process.returncode != 0:
            return None

        public_ip = stdout.decode().strip()
        return public_ip if public_ip != 'None' else None

    def get_tasks_by_type(self, tasks: List[ECSTask], task_type: str) -> List[ECSTask]:
        """Filter tasks by type"""
        return [task for task in tasks if task.task_type == task_type]
    
    async def get_mapreduce_ips(self) -> TaskIPs:
        """Get IPs for splitter, mappers, and reducer"""
        tasks = await self.get_all_tasks()
        
        splitters = self.get_tasks_by_type(tasks, "splitter")
        mappers = self.get_tasks_by_type(tasks, "mapper")
        reducers = self.get_tasks_by_type(tasks, "reducer")
        
        if not splitters:
            raise Exception("No splitter tasks found")
        if not mappers:
            raise Exception("No mapper tasks found")
        if not reducers:
            raise Exception("No reducer tasks found")
        
        return TaskIPs(
            splitter_ip=splitters[0].public_ip,
            mapper_ips=[mapper.public_ip for mapper in mappers],
            reducer_ip=reducers[0].public_ip
        )
    
    async def clean_s3_results(self):
        import subprocess
        subprocess.run([
            "aws", "s3", "rm", f"s3://{self.s3_bucket}/map_results/", 
            "--recursive", "--region", "us-west-2"
        ], check=False) 
        subprocess.run([
            "aws", "s3", "rm", f"s3://{self.s3_bucket}/chunks/", 
            "--recursive", "--region", "us-west-2"
        ], check=False)

if __name__ == "__main__":
    async def main():
        manager = ECSTaskManager(cluster_name="hello-cluster", region="us-west-2")
        tasks = await manager.get_all_tasks()
        for task in tasks:
            print(f"Task ARN: {task.task_arn}, Type: {task.task_type}, IP: {task.public_ip}, Status: {task.status}")
        
        try:
            ips = await manager.get_mapreduce_ips()
            print(f"Splitter IP: {ips.splitter_ip}")
            print(f"Mapper IPs: {', '.join(ips.mapper_ips)}")
            print(f"Reducer IP: {ips.reducer_ip}")
        except Exception as e:
            print(f"Error fetching MapReduce IPs: {e}")

    asyncio.run(main())