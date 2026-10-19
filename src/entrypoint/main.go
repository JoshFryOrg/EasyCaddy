// The image's entrypoint. The image has no shell, so this does what a shell script would: tidy the Caddyfile's
// formatting, then become Caddy. Standard library only, built static.
//
// With arguments it runs them instead, so `docker run <image> caddy hash-password` and
// `docker run <image> caddy validate --config /etc/caddy/Caddyfile` work.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

const (
	caddy     = "/usr/bin/caddy"
	caddyfile = "/etc/caddy/Caddyfile"
)

func main() {
	args := os.Args[1:]
	if len(args) > 0 {
		run(args)
	}

	// A read-only or missing Caddyfile must not stop Caddy starting, so formatting is best-effort.
	fmt.Println("Formatting Caddyfile")
	format := exec.Command(caddy, "fmt", "--overwrite", caddyfile)
	format.Stdout, format.Stderr = os.Stdout, os.Stderr
	if err := format.Run(); err != nil {
		fmt.Println("Caddyfile not formatted, continuing")
	}

	fmt.Println("Launching Caddy")
	run([]string{caddy, "run", "--config", caddyfile, "--adapter", "caddyfile"})
}

// run replaces this process with the command, so it gets the container's signals and its exit code is the
// container's. `caddy` names the image's Caddy, as the image's PATH has nothing else.
func run(args []string) {
	path := args[0]
	if path == "caddy" {
		path = caddy
	} else if found, err := exec.LookPath(path); err == nil {
		path = found
	}
	err := syscall.Exec(path, args, os.Environ())
	fmt.Fprintf(os.Stderr, "easycaddy: cannot run %s: %v\n", args[0], err)
	os.Exit(127)
}
