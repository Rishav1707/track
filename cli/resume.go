package cli

import (
	"strings"
	"time"

	"github.com/mlange-42/track/core"
	"github.com/mlange-42/track/out"
	"github.com/spf13/cobra"
)

func resumeCommand(t *core.Track) *cobra.Command {
	start := &cobra.Command{
		Use:     "resume [message]",
		Short:   "Resume the last project",
		Aliases: []string{"re"},
		Args:    cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			last, err := t.LatestRecord()
			if err != nil {
				out.Err("failed to resume: %s", err)
				return
			}
			if !last.HasEnded() {
				out.Err("failed to resume: record running in '%s'", last.Project)
				return
			}

			project := last.Project

			if !t.ProjectExists(project) {
				out.Err("failed to resume: project '%s' does not exist", project)
				return
			}

			note := last.Note
			tags := last.Tags
			if len(args) > 0 {
				note = strings.Join(args[1:], " ")
				tags = t.ExtractTags(args[1:])
			}

			record, err := t.StartRecord(project, note, tags, time.Now())
			if err != nil {
				out.Err("failed to resume: %s", err.Error())
				return
			}

			out.Success("Resume record in '%s' at %02d:%02d", project, record.Start.Hour(), record.Start.Minute())
		},
	}

	return start
}
