<img width="1210" height="645" alt="Screenshot 2025-09-25 at 13 37 01" src="https://github.com/user-attachments/assets/1040cac4-199a-4df0-b79e-4514e94e4e53" /># Homework4
## Part 1: Reading
### Most Informative
The most enlightening concept was how RPC extends local procedure calling to distributed environments:  
"It is a powerful technique based on extending the notion of local procedure calling, so that the called procedure may not exist in the same address space as the calling procedure."  
This abstraction makes distributed computing feel familiar while hiding the complexity underneath.
### What you were already experienced with, and where you got that experience!
TCP and UDP protocal knowledge, I learn it from the Network course.
## Part 2: Infrastructure setup
**Task running**
<img width="1210" height="645" alt="Screenshot 2025-09-25 at 13 37 01" src="https://github.com/user-attachments/assets/7804e86c-6fcc-4282-bdc9-ef69efaecbad" />
**Service connected**
<img width="725" height="319" alt="image" src="https://github.com/user-attachments/assets/261eff55-c604-4d21-a858-d13e539fda3a" />
### EC2 vs. ECS
EC2 is a system allowing customers to manage virtual machines. ECS is a service to manage containers.
### VPC & Subnet
VPC is a kind of virtual network implementation, enabling users to launch AWS resources in a private virtual network. A subnet is a range of IP addresses within your VPC. It's a way to partition your VPC's IP address space.
### TCP vs. UDP
TCP guarantees the packets to be sent to the receiver by 3-way handshaking, ACK, retransimission.  
**However**, modern web applications are prone to use UDP by introducing QUIC to take advantages of TCP, such as HTTP/3.
