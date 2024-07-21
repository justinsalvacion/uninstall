# Uninstaller Tool

This Go program allows you to uninstall programs from a Windows system using the Windows registry. It supports both MSI and non-MSI installers.

## Features

- Finds programs installed on a Windows system by querying the registry.
- Lists programs matching a given name.
- Uninstalls selected program using the appropriate method (MSI or non-MSI).


## Usage

1. Run the executable:
    ```bash
    ./uninstaller
    ```

2. Enter the program name when prompted:
    ```
    Enter the program name: [your program name]
    ```

3. Select the program to uninstall from the list:
    ```
    Select a program to uninstall:
    1: Example Program 1 (Standard)
    2: Example Program 2 (Wow6432Node)
    Enter selection number: [your selection]
    ```

4. For non-MSI uninstallers, enter the silent uninstall parameter when prompted:
    ```
    Enter the silent uninstall parameter: [your parameter]
    ```

