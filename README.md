# CPU Frequency Manager (Linux)
Fluctuates CPU Frequency to match CPU's Work while keeping the CPU at a reasonable temperature. 

## How it Works
The program starts by gaining access to the CPU's maximum frequency. Then, sets the CPU at a low frequency based on the Max Frequency. When the CPU recieves work, frequency is spiked for a set amount of time in order to max given work.

## How to Run
You **HAVE** to run with escalated privileges!

**Dependencies**
- `cpupower`: Used to manage CPU Frequency
- `sensors`: Used to obtain CPU Temperatures
- `golang`: Language used to run program

**Running Program (No Output)**
```bash
# From Root Directory of Repository
sudo go run ./src
```

**Running Program (With Output)**
```bash
# From Root Directory of Repository
sudo go run ./src -d        # Run in Debug Mode
```

## License
Licensed under [MIT](LICENSE).