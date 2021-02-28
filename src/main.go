package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// CPU Struct
type CPU struct {
	cpu []string
}

func (c *CPU) parse(output []byte) (sum, work int) {
	outStr := string(output)

	// SPLIT CPU
	i := strings.Index(outStr, "\n")
	c.cpu = strings.SplitAfter(outStr[:i], " ")

	// SUM USAGE
	totalJiffies := 0
	for _, val := range c.cpu[2:] {
		i, _ := strconv.Atoi(strings.TrimRight(val, " "))
		totalJiffies += i
	}

	workJiffies := 0
	for _, val := range c.cpu[2:5] {
		i, _ := strconv.Atoi(strings.TrimRight(val, " "))
		workJiffies += i
	}

	return totalJiffies, workJiffies
}

func handleError(e error) {
	if e != nil {
		panic(e)
	}
}

func getCPUOutput() []byte {
	cmd := exec.Command("cat", "/proc/stat")
	out, err := cmd.Output()
	handleError(err)

	return out
}

func setCPUFreq(freq string) {
	cmd := exec.Command("cpupower", "frequency-set", "-u", freq)
	_, err := cmd.Output()
	handleError(err)
}

// TempInfo - Simple CPU Temperature Information Strucutre
type TempInfo struct {
	packageTemp float64
	coreTemps   []float64
}

// Print all Temperature Information
func (tInfo *TempInfo) print() {
	fmt.Printf("Package Temp [%.2f]\n", tInfo.packageTemp)
	for i := 0; i < len(tInfo.coreTemps); i++ {
		fmt.Printf("Core%d [%.2f]\n", i, tInfo.coreTemps[i])
	}
}

// Parses Output into Object
func parseOutput(output []byte, tInfo *TempInfo) {
	strOut := string(output)

	// GET PACKAGE SECTION
	index := strings.Index(strOut, "Package")
	endIndex := index + (strings.Index(strOut[index:], "Core"))
	pkgSection := strOut[index:endIndex]

	// // PACKAGE TEMP
	index += strings.Index(pkgSection, "input")
	endIndex = index + (strings.Index(strOut[index:], "\n"))
	tempArr := strings.Split(strOut[index:endIndex], " ")[1]
	val, e := strconv.ParseFloat(tempArr, 64)
	handleError(e)
	tInfo.packageTemp = val

	// // GET CORE TEMP SECTION
	foundAllCores := false
	var coreSection string
	tInfo.coreTemps = make([]float64, 0, 10) // Allocate 10 Spots

	for !foundAllCores {
		// GET SECTION
		index = endIndex + strings.Index(strOut[endIndex:], "Core")
		endIndex = index + strings.Index(strOut[index+1:], "Core")

		// VERIFY FOUND
		if endIndex > index {
			coreSection = strOut[index:endIndex] // Still More
		} else {
			coreSection = strOut[index:] // Found Last Core
			foundAllCores = true
		}

		// PARSE TEMP
		i1 := strings.Index(coreSection, "input")
		i2 := i1 + strings.Index(coreSection[i1:], "\n")
		temp, err := strconv.ParseFloat(strings.Split(coreSection[i1:i2], " ")[1], 64)
		handleError(err)

		// STORE TEMP
		tInfo.coreTemps = append(tInfo.coreTemps, temp)
	}
}

// Obtains Package temperature and returns it
func getPackageTemp() float64 {
	var tempInfo TempInfo
	cmd := exec.Command("sensors", "-u")
	out, err := cmd.Output()
	handleError(err)

	// Parse the Output and Obtain Temp Info
	parseOutput(out, &tempInfo)

	// Return the Package's Temp
	return tempInfo.packageTemp
}

// Obtains the Maximum Frequency in GHz for CPU
// @returns Floating Point Value
func getMaxFreq() float64 {
	dat, err := ioutil.ReadFile("/sys/bus/cpu/devices/cpu0/cpufreq/cpuinfo_max_freq")
	handleError(err)
	val, err := strconv.ParseFloat(strings.Replace(string(dat), "\n", "", -1), 64)
	handleError(err)
	return val / 1000000
}

func main() {
	// ARGUMENT CONFIG
	isDebug := false
	for _, arg := range os.Args {
		if arg == "-d" { // Switch Debug ON
			println("Debug Mode ON")
			isDebug = true
		}
	}

	// CONFIG USED
	interval := 1 * time.Second // Seconds
	boostTimer := -1            // Initiate the Boost Timer
	currFreq := 0.0             // Keep track of Current Frequency
	maxFreq := getMaxFreq()     // Get the Max CPU Frequency

	// START RUNNING
	cpu := CPU{}
	// var tJ1, tJ2, tW1, tW2 int
	var tJ1, tJ2, tW1, tW2 int

	// INITIAL VALUES
	tJ1, tW1 = cpu.parse(getCPUOutput())

	for {
		time.Sleep(interval)                 // Every Interval
		tJ2, tW2 = cpu.parse(getCPUOutput()) // Get Current Data

		// Calculate Usage
		dWork := tW2 - tW1
		dJiff := tJ2 - tJ1

		// If Monitor is off (Assume Idle)
		// Check if Boosting
		if boostTimer == -1 { // No Boost
			if dWork >= 200 { // Heavy Load
				// Obtain CPU Temp
				cpuTemp := getPackageTemp()

				// Check Temperature to set Boost Timer
				if cpuTemp < 60.00 && currFreq != 3.1 {
					freq := fmt.Sprintf("%.2fGHZ", maxFreq)

					if isDebug {
						fmt.Printf("CPU Freq Set to '%s'\n", freq)
					}
					setCPUFreq(freq)

					currFreq = maxFreq
					boostTimer = 2 // 2 Seconds
				} else if cpuTemp < 70.00 && currFreq != 2.8 { // CPU is HOT
					freq := fmt.Sprintf("%.2fGHZ", maxFreq-0.3)

					if isDebug {
						fmt.Printf("CPU Freq Set to '%s'\n", freq)
					}
					setCPUFreq(freq)

					currFreq = maxFreq - 0.3
					boostTimer = 1
				} else if currFreq != 2.6 { // Cool CPU Down!
					freq := fmt.Sprintf("%.2fGHZ", maxFreq-0.5)

					if isDebug {
						fmt.Printf("CPU Freq Set to '%s'\n", freq)
					}
					setCPUFreq(freq)

					currFreq = maxFreq - 0.5
					boostTimer = 5
				}

				if isDebug {
					fmt.Printf("Boost Init = %d\n", boostTimer)
				}

			} else if dWork > 100 && currFreq != 2.5 { // Medium Load
				freq := fmt.Sprintf("%.2fGHZ", maxFreq-0.6)

				if isDebug {
					fmt.Printf("CPU Freq Set to '%s'\n", freq)
				}
				setCPUFreq(freq)
				currFreq = maxFreq - 0.6
			} else if currFreq != 2.25 { // Idle
				freq := fmt.Sprintf("%.2fGHZ", maxFreq-0.85)

				if isDebug {
					fmt.Printf("CPU Freq Set to '%s'\n", freq)
				}
				setCPUFreq(freq)
				currFreq = maxFreq - 0.85
			}

		} else {
			// Decrement Boost Timer
			boostTimer--
			if isDebug {
				fmt.Printf("Boost Timer Decrement: %d\n", boostTimer)
			}
		}

		if isDebug {
			fmt.Printf("dWork: %d\n", dWork)
			fmt.Printf("dJiff: %d\n\n", dJiff)
		}

		// Store Previous Values
		tW1 = tW2
		tJ1 = tJ2
	}
}
