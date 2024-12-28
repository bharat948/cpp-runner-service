package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

type Request struct {
    Code      string `json:"code" binding:"required"`
    Language  string `json:"language" binding:"required"`
    TestInput string `json:"test_input" binding:"required"`
}

type Response struct {
    Output       string `json:"output"`
    Error        string `json:"error"`
    MemoryUsage  string `json:"memory_usage"`
    CPUUsage     string `json:"cpu_usage"`
    CompileError string `json:"compile_error,omitempty"`
    RuntimeError string `json:"runtime_error,omitempty"`
}

func main() {
    r := gin.Default()

    r.POST("/run", func(c *gin.Context) {
        var req Request
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(400, gin.H{"error": "Invalid request format"})
            return
        }

        if req.Language != "cpp" {
            c.JSON(400, gin.H{"error": "Unsupported language"})
            return
        }

        response, err := executeCode(req)
        if err != nil {
            c.JSON(500, Response{
                Error: fmt.Sprintf("Execution failed: %v", err),
            })
            return
        }

        c.JSON(200, response)
    })

    r.Run(":8080")
}

func executeCode(req Request) (*Response, error) {
    // Create temporary directory for compilation
    tempDir, err := ioutil.TempDir("", "cpp-execution")
    if err != nil {
        return nil, fmt.Errorf("failed to create temp directory: %v", err)
    }
    defer os.RemoveAll(tempDir)

    // Write code to file
    sourceFile := fmt.Sprintf("%s/main.cpp", tempDir)
    if err := ioutil.WriteFile(sourceFile, []byte(req.Code), 0644); err != nil {
        return nil, fmt.Errorf("failed to write source file: %v", err)
    }

    // Compile the code
    execFile := fmt.Sprintf("%s/main", tempDir)
    compileOutput, compileErr := compileCode(sourceFile, execFile)
    if compileErr != nil {
        return &Response{
            CompileError: compileOutput,
            Error:       "Compilation failed",
        }, nil
    }

    // Run the compiled code
    return runCompiledCode(execFile, req.TestInput)
}

func compileCode(sourceFile, execFile string) (string, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    cmd := exec.CommandContext(ctx, "g++", sourceFile, "-o", execFile, "-Wall")
    var stderr bytes.Buffer
    cmd.Stderr = &stderr

    if err := cmd.Run(); err != nil {
        return stderr.String(), err
    }

    return "", nil
}

func runCompiledCode(execFile, input string) (*Response, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    cmd := exec.CommandContext(ctx, execFile)
    
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
    
    if input != "" {
        cmd.Stdin = bytes.NewBufferString(input)
    }

    start := time.Now()
    runErr := cmd.Run()
    elapsed := time.Since(start)

    // Get memory stats
    var memStats runtime.MemStats
    runtime.ReadMemStats(&memStats)

    response := &Response{
        Output:      stdout.String(),
        MemoryUsage: fmt.Sprintf("%v KB", memStats.Alloc/1024),
        CPUUsage:    fmt.Sprintf("%v ms", elapsed.Milliseconds()),
    }

    if runErr != nil {
        if ctx.Err() == context.DeadlineExceeded {
            response.RuntimeError = "Execution timeout"
        } else {
            response.RuntimeError = stderr.String()
        }
        response.Error = "Runtime error occurred"
    }

    return response, nil
}