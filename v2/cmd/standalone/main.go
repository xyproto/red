package main

import (
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/Netflix/go-expect"
	"golang.org/x/crypto/ssh/terminal"
)

// ctrl returns the control character itself based on the input letter.
func ctrl(letter rune) string {
	return string(letter & 0x1F)
}

func main() {
	// Initialize the expect console with stdout
	c, err := expect.NewConsole(expect.WithStdout(os.Stdout))
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	// Set terminal to raw mode to ensure proper control character handling
	fd := int(c.Tty().Fd())
	oldState, err := terminal.MakeRaw(fd)
	if err != nil {
		log.Fatal(err)
	}
	defer terminal.Restore(fd, oldState) // Ensure we restore on exit

	// Execute the command
	cmd := exec.Command("o", "/tmp/ost")
	cmd.Stdin = c.Tty()
	cmd.Stdout = c.Tty()
	cmd.Stderr = c.Tty()

	go func() {
		c.ExpectEOF() // Wait for command to finish
	}()

	// Start the command and handle errors
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	// Send "hello" with a slight delay to ensure it’s processed
	time.Sleep(500 * time.Millisecond)
	c.Send("hello")
	time.Sleep(500 * time.Millisecond)

	// Send Ctrl-S to save
	c.Send(ctrl('s'))
	time.Sleep(500 * time.Millisecond)

	// Send Ctrl-Q to quit
	c.Send(ctrl('q'))
	time.Sleep(500 * time.Millisecond)

	// Wait for the command to finish execution
	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}
}
