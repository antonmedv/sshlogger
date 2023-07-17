package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/gliderlabs/ssh"
)

func main() {
	ssh.Handle(func(s ssh.Session) {
		echo := func(msg string) { s.Write([]byte(msg + "\n")) }
		exit := func(code int) { s.Exit(code) }

		rawUser := s.User()
		rawCommand := s.RawCommand()
		hostname, _ := os.Hostname()

		fmt.Printf("%s@%s$ %s\n", rawUser, hostname, rawCommand)
		cmd := exec.Command("bash", "-c", rawCommand)
		cmd.Env = os.Environ()
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			echo("error: failed to create stdout pipe")
			exit(1)
			return
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			echo("error: failed to create stderr pipe")
			exit(1)
			return
		}
		if err := cmd.Start(); err != nil {
			echo("error: failed to start command")
			exit(1)
			return
		}
		go io.Copy(s, stdout)
		go io.Copy(s.Stderr(), stderr)

		// Wait for the command to finish.
		exitStatus := 0
		if err := cmd.Wait(); err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				exitStatus = exitError.ExitCode()
			} else {
				echo("error: failed to wait for command")
				exit(1)
				return
			}
		}
		exit(exitStatus)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "2222"
	}
	log.Fatal(ssh.ListenAndServe(":"+port, nil))
}
