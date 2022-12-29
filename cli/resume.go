package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/mlange-42/track/core"
	"github.com/mlange-42/track/out"
	"github.com/mlange-42/track/util"
	"github.com/spf13/cobra"
)

func resumeCommand(t *core.Track) *cobra.Command {
	var useLast bool
	var skip bool

	resume := &cobra.Command{
		Use:   "resume [NOTE...]",
		Short: "Resume a paused or stopped project",
		Long: `Resume a paused or stopped project

The note argument provides a note for the pause when resuming a stopped record`,
		Aliases: []string{"re"},
		Args:    util.WrappedArgs(cobra.ArbitraryArgs),
		Run: func(cmd *cobra.Command, args []string) {
			open, err := t.OpenRecord()
			if err != nil {
				out.Err("failed to resume: %s", err.Error())
				return
			}
			if open != nil {
				if useLast {
					out.Err("failed to resume: Flag --last not permitted when resuming a running record")
					return
				}
				if len(args) > 0 {
					out.Err("failed to resume: no positional arguments accepted when resuming a running record")
					return
				}

				pause, err := resumeOpenRecord(t, open, skip)
				if err != nil {
					out.Err("failed to resume: %s", err.Error())
					return
				}
				skipped := ""
				if skip {
					skipped = fmt.Sprintf(" (skipped %s pause)", util.FormatDuration(pause))
				}
				out.Success("Resume record in '%s'%s", open.Project, skipped)
				return
			}

			last, err := t.LatestRecord()
			if err != nil {
				out.Err("failed to resume: %s", err.Error())
				return
			}
			if last == nil {
				out.Err("failed to resume: no record found")
				return
			}
			if !useLast {
				out.Err("failed to resume: no running record. To resume a previous record, use --last")
				return
			}

			pause, err := resumeLastRecord(t, last, args, skip)
			if err != nil {
				out.Err("failed to resume: %s", err.Error())
				return
			}
			skipped := ""
			if skip {
				skipped = fmt.Sprintf(" (skipped %s pause)", util.FormatDuration(pause))
			}
			out.Success("Resume record in '%s'%s", last.Project, skipped)
		},
	}

	resume.Flags().BoolVarP(&useLast, "last", "l", false, "Continue the last record instead of a running one")
	resume.Flags().BoolVarP(&skip, "skip", "s", false, "Resume, and delete the running pause/gap")

	return resume
}

func resumeOpenRecord(t *core.Track, open *core.Record, skip bool) (time.Duration, error) {
	if !open.IsPaused() {
		return 0, fmt.Errorf("record is not paused")
	}
	var duration time.Duration
	if skip {
		pause, _ := open.PopPause()
		duration = pause.Duration(time.Time{}, time.Time{})
	} else {
		_, err := open.EndPause(time.Now())
		if err != nil {
			return 0, err
		}
	}
	return duration, t.SaveRecord(open, true)
}

func resumeLastRecord(t *core.Track, last *core.Record, args []string, skip bool) (time.Duration, error) {
	project := last.Project

	if !t.ProjectExists(project) {
		return 0, fmt.Errorf("project '%s' does not exist", project)
	}
	proj, err := t.LoadProjectByName(project)
	if err != nil {
		return 0, err
	}
	if proj.Archived {
		return 0, fmt.Errorf("project '%s' is archived", proj.Name)
	}

	now := time.Now()

	oldEnd := last.End
	last.End = time.Time{}

	var duration time.Duration
	if skip {
		duration = now.Sub(oldEnd)
	} else {
		pause, err := last.InsertPause(oldEnd, now, strings.Join(args, " "))
		if err != nil {
			return 0, err
		}
		duration = pause.Duration(time.Time{}, time.Time{})
	}

	return duration, t.SaveRecord(last, true)
}
