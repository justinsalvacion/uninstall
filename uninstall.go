package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"golang.org/x/sys/windows/registry"
)

const (
	uninstallKeyPath           = `SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`
	uninstallKeyPathWow6432Node = `SOFTWARE\Wow6432Node\Microsoft\Windows\CurrentVersion\Uninstall`
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter the program name: ")
	programName, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Error reading program name: %v", err)
	}
	programName = strings.TrimSpace(programName)

	programs, err := findPrograms(programName)
	if err != nil {
		log.Fatalf("Error finding programs: %v", err)
	}

	if len(programs) == 0 {
		fmt.Println("No programs found.")
		return
	}

	fmt.Println("Select a program to uninstall:")
	for i, program := range programs {
		fmt.Printf("%d: %s (%s)\n", i+1, program.DisplayName, program.RegistryLocation)
	}

	var selection int
	fmt.Print("Enter selection number: ")
	_, err = fmt.Scan(&selection)
	if err != nil || selection < 1 || selection > len(programs) {
		log.Fatalf("Invalid selection: %v", err)
	}

	// added thie line to fix a bug bc apparently when u subsequently read from the buffer, it captures the leftover newline character.
	reader.ReadString('\n')

	selectedProgram := programs[selection-1]
	if strings.HasPrefix(strings.ToLower(selectedProgram.UninstallString), "msiexec") {
		err = uninstallMSI(selectedProgram)
	} else {
		err = uninstallNonMSI(selectedProgram)
	}

	if err != nil {
		log.Fatalf("Uninstall failed: %v", err)
	} else {
		fmt.Println("Uninstall completed successfully.")
	}
}

type Program struct {
	DisplayName     string
	UninstallString string
	RegistryLocation string
}

func findPrograms(programName string) ([]Program, error) {
	programs, err := findProgramsInKey(uninstallKeyPath, programName, "Standard")
	if err != nil {
		return nil, err
	}

	programsWow6432Node, err := findProgramsInKey(uninstallKeyPathWow6432Node, programName, "Wow6432Node")
	if err != nil {
		return nil, err
	}

	programs = append(programs, programsWow6432Node...)
	return programs, nil
}

func findProgramsInKey(registryPath, programName, registryLocation string) ([]Program, error) {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, registryPath, registry.READ)
	if err != nil {
		return nil, err
	}
	defer key.Close()

	var programs []Program
	subKeyNames, err := key.ReadSubKeyNames(-1)
	if err != nil {
		return nil, err
	}

	for _, subKeyName := range subKeyNames {
		subKey, err := registry.OpenKey(key, subKeyName, registry.READ)
		if err != nil {
			continue
		}
		defer subKey.Close()

		displayName, _, err := subKey.GetStringValue("DisplayName")
		if err != nil {
			continue
		}

		if strings.Contains(strings.ToLower(displayName), strings.ToLower(programName)) {
			uninstallString, _, err := subKey.GetStringValue("UninstallString")
			if err != nil {
				continue
			}
			programs = append(programs, Program{DisplayName: displayName, UninstallString: uninstallString, RegistryLocation: registryLocation})
		}
	}
	return programs, nil
}

func uninstallMSI(program Program) error {
	uninstallString := strings.ReplaceAll(program.UninstallString, "/I", "/X")
	uninstallString = strings.ReplaceAll(uninstallString, "/i", "/X")

	cmdArgs := strings.Split(uninstallString, " ")
	cmdArgs = append(cmdArgs, "/qn", "/norestart")
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	fmt.Printf("Executing command: %s %s\n", cmdArgs[0], strings.Join(cmdArgs[1:], " "))
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Uninstall command failed: %s\nOutput: %s", err, output)
		return err
	}
	fmt.Printf("Uninstall output: %s", output)
	return nil
}

func uninstallNonMSI(program Program) error {
	reader := bufio.NewReader(os.Stdin)
	var silentParam string
	for silentParam == "" {
		fmt.Print("Enter the silent uninstall parameter: ")
		silentParam, _ = reader.ReadString('\n')
		silentParam = strings.TrimSpace(silentParam)
		if silentParam == "" {
			fmt.Println("Silent uninstall parameter cannot be empty. Please try again.")
		}
	}
	fmt.Printf("Silent uninstall parameter received: %s\n", silentParam)

	uninstallString := program.UninstallString
	if strings.HasPrefix(uninstallString, "\"") && strings.HasSuffix(uninstallString, "\"") {
		uninstallString = uninstallString[1 : len(uninstallString)-1]
	}

	uninstallPath, err := exec.LookPath(uninstallString)
	if err != nil {
		log.Fatalf("Uninstall command not found: %v", err)
	}

	cmdArgs := append([]string{uninstallPath}, strings.Split(silentParam, " ")...)
	fmt.Printf("Executing command: %s %s\n", cmdArgs[0], strings.Join(cmdArgs[1:], " "))
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Uninstall command failed: %s\nOutput: %s", err, output)
		return err
	}
	fmt.Printf("Uninstall output: %s", output)
	return nil
}
