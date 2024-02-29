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

	fmt.Println("Loading WDS-Image-Replace v1.3")
	exe, err := os.Executable()
    if err != nil {
        panic(err)
    }

    root = filepath.Dir(exe)
	fmt.Println("Loaded in "+root)
	fmt.Println("Get config")
	config = getConfig()
	
	PS_Script()
	
}

func PS_Script(){
	fmt.Println("Checking config")
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
	fmt.Println("Starting Image replace procces this can take a few minutes")
    
	// Start the timer
	done := make(chan bool)
	var elapsedTime int32

	go func() {
		elapsedTime = startTimer(done)
	}()

	

	psScript := `
        $ErrorActionPreference = "Stop" # Make sure any error is treated as a terminating error

        try {
			WDSUTIL /Replace-Image /Image:"`+config.Image.Name+`" /ImageType:Install /ImageGroup:"`+config.Image.Group+`" /ReplacementImage /ImageFile:"`+config.Image.Path+`" /Name:"`+config.Image.Name+`" /Server:`+config.Server+`
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
			fmt.Println("WDS-Image-replace encounterd a error. Check the log file and please try again.")
			exitErr, ok := err.(*exec.ExitError)
			if ok {
				// The command failed to execute properly
				logger("Error PowerShell script failed with exit code: " + fmt.Sprint(exitErr.ExitCode()))
				if stderrBuf.Len() > 0 {
					logger("PowerShell error output: " + stderrBuf.String())
				}
			} else {
				// Some other error occurred
				logger("Error executing PowerShell script: " + err.Error())
			}
			return
		}
	
		done <- true

		fmt.Println("Image succesfully replaced in "+fmt.Sprint(elapsedTime)+" seconds!")
		logger("Image succesfully replaced in "+fmt.Sprint(elapsedTime)+" seconds!")

		return
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
    _, err = writer.WriteString(timeStamp() + ": "+toASCII(msg)+"\n") // The \n is the newline character
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


func toASCII(s string) string {
    var asciiStr string
    for _, r := range s {
        if r <= 127 {
            asciiStr += string(r)
        } else {
            // Replace non-ASCII character with '?'
            // or you can just skip it with `continue`
            asciiStr += "?"
        }
    }
    return asciiStr
}

func startTimer(done <-chan bool) int32 {
	seconds := 0

	for {
		select {
		case <-done:
			return int32(seconds)
		default:
			fmt.Printf("\rElapsed time: %d seconds", seconds)
			time.Sleep(time.Second)
			seconds++
		}
	}
}