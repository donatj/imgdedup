package main

import (
	"log"
	"os/exec"
	"time"
)

func diffToolDiff(tool string, diffs []ImgDiff) {
	for _, diff := range diffs {
		if diff.Diff > 0 || diff.Left.Filesize != diff.Right.Filesize {
			if *difftool != "" {
				diffTool(tool, diff.Left.Path, diff.Right.Path)
				time.Sleep(500 * time.Millisecond)
			}
		}
	}
}

func diffTool(tool string, leftf string, rightf string) {
	log.Println("Launching difftool")
	cmd := exec.Command(tool, leftf, rightf)
	cmd.Run()
}
