package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"os"
	"bufio"
	"time"
	"io"
	"encoding/json"
	"path/filepath"
)

type Image struct {
    Group	string ``
    Name   	string ``
    Path 	string ``
}

type Log struct {
    Name   	string ``
    Path 	string ``
}

type Config struct {
    Server	string ``
    Image	Image  ``
    Log		Log    ``
}

var root string = ""
var config Config 

func main() {

	exe, err := os.Executable()
    if err != nil {
        panic(err)
    }

    root = filepath.Dir(exe)
	config = getConfig()

	PS_Script()

}

func PS_Script(){

	configError := false

	if(config.Server == ""){
		logger("Error WDS Server not defined in config")
		configError = true
	}

	if(config.Image.Group == ""){
		logger("Error WDS Install image group not defined in config")
		configError = true
	}

	if(config.Image.Name == ""){
		logger("Error WDS Install image name not defined in config")
		configError = true
	}

	if(!checkFile(config.Image.Path)){
		logger("Error image file does not exist at: "+config.Image.Path)
		configError = true
	}

	if(configError){
		return
	}

    // PowerShell script as a string
    psScript := `
        $ErrorActionPreference = "Stop" # Make sure any error is treated as a terminating error
        $WDSserver = "` + config.Server + `"
        $ImageGroup = "` + config.Image.Group + `"
        $ImageName = "` + config.Image.Name + `"
        $NewImageFilePath = "` + config.Image.Path + `"

        try {
            Import-Module WdsMgmt
            Replace-WdsInstallImage -Server $WDSserver -ImageGroup $ImageGroup -ImageName $ImageName -ReplacementImagePath $NewImageFilePath
        } catch {
            Write-Error $_.Exception.Message
            exit 1
        }
    `

    // Run PowerShell command
    cmd := exec.Command("powershell", "-command", psScript)
    var stdoutBuf, stderrBuf bytes.Buffer
    cmd.Stdout = &stdoutBuf
    cmd.Stderr = &stderrBuf
    err := cmd.Run()

    if err != nil {
        logger("Error executing PowerShell script: " + err.Error())
        if stderrBuf.Len() > 0 {
            logger("PowerShell error output: " + stderrBuf.String())
        }
        return
    }

    if stdoutBuf.Len() > 0 {
        logger("PowerShell output: " + stdoutBuf.String())
    }

}

func getConfig() Config{

	filePath := root+"\\config.json"

    // Open the JSON file
    file, err := os.Open(filePath)
  
	if err != nil {
        logger("Error opening JSON file:"+filePath)
        panic(err)
    }
    defer file.Close()

	bytes, err := io.ReadAll(file)
	
    // Read the file content into a byte slice
    if err != nil {
        logger("Error reading "+filePath+":"+ err.Error())
        panic(err)
    }

	configJSON := string(bytes)
	
	var config Config
    // Unmarshal the byte slice into the struct
    err = json.Unmarshal([]byte(configJSON), &config)
    if err != nil {
		logger("Error unmarshalling "+filePath+":"+ err.Error())
        panic(err)
    }
	
	return config
}

func logger(msg string){

	filePath := root

	if(config.Log.Path != ""){
		filePath = config.Log.Path
	}
    
	if(config.Log.Name == ""){
		filePath = filePath+"\\log.txt"
	}else{
		filePath = filePath+"\\"+config.Log.Name
	}

    // Open the file in append mode. If the file doesn't exist, it will be created
    file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        fmt.Println("Error opening the file:", err)
        return
    }
    defer file.Close()

    // Create a writer for the file
    writer := bufio.NewWriter(file)

    // Write a new line to the file
    _, err = writer.WriteString(timeStamp() + ": "+msg+"\n") // The \n is the newline character
    if err != nil {
        fmt.Println("Error writing to the file:", err)
        return
    }

    // Flush the buffer to ensure all data is written to the file
    err = writer.Flush()
    if err != nil {
        fmt.Println("Error flushing the buffer:", err)
        return
    }

}

func checkFile(path string)bool{

	if _, err := os.Stat(path); os.IsNotExist(err) {
        return false
    } else {
        return true
    }

}

func timeStamp() string{
	time := time.Now().Format("2006-01-02 15:04:05")
	return time
}