# CPU Frequency Manager (Linux)
Fluctuates CPU Frequency to match CPU's Work while keeping the CPU at a reasonable temperature. 

## How it Works
The program starts by gaining access to the CPU's maximum frequency. Then, sets the CPU at a low frequency based on the Max Frequency. When the CPU recieves work, frequency is spiked for a set amount of time in order to max given work.

## How to Run
**Dependencies**
- `cpupower`: Used to manage CPU Frequency
- `golang`: Language used to run program

**Running Program**

You **HAVE** to run with escalated privileges!
```bash
# From Root Directory of Repository
sudo go run ./src
```