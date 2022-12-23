package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/mlange-42/track/core"
	"github.com/mlange-42/track/fs"
	"github.com/mlange-42/track/out"
	"github.com/mlange-42/track/util"
	"github.com/spf13/cobra"
)

var (
	// ErrUserAbort is an error for abort by the user
	ErrUserAbort = errors.New("aborted by user")
)

func editCommand(t *core.Track) *cobra.Command {
	create := &cobra.Command{
		Use:     "edit",
		Short:   "Edit a resource",
		Aliases: []string{"e"},
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	create.AddCommand(editProjectCommand(t))
	create.AddCommand(editRecordCommand(t))

	return create
}

func editRecordCommand(t *core.Track) *cobra.Command {
	editProject := &cobra.Command{
		Use:     "record <TIME>",
		Short:   "Edit a record",
		Aliases: []string{"r"},
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			timeString := strings.Join(args, " ")
			tm, err := util.ParseDateTime(timeString)
			if err != nil {
				out.Err("failed to edit project: %s", err)
				return
			}
			err = editRecord(t, tm)
			if err != nil {
				out.Err("failed to edit project: %s", err)
				return
			}
			out.Success("Saved record '%s'", tm.Format(util.DateTimeFormat))
		},
	}

	return editProject
}

func editProjectCommand(t *core.Track) *cobra.Command {
	editProject := &cobra.Command{
		Use:     "project <NAME>",
		Short:   "Edit a project",
		Aliases: []string{"p"},
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			err := editProject(t, name)
			if err != nil {
				out.Err("failed to edit project: %s", err)
				return
			}
			out.Success("Saved project '%s'", name)
		},
	}

	return editProject
}

func editRecord(t *core.Track, tm time.Time) error {
	record, err := t.LoadRecordByTime(tm)
	if err != nil {
		return err
	}

	return edit(t, &record, func(b []byte) error {
		var newRecord core.Record
		if err := json.Unmarshal(b, &newRecord); err != nil {
			return err
		}

		// TODO could change, but requires deleting original file
		if newRecord.Start != record.Start {
			return fmt.Errorf("can't change start time")
		}

		if !newRecord.End.IsZero() && newRecord.End.Before(newRecord.Start) {
			return fmt.Errorf("end time is before start time")
		}

		if err = t.SaveRecord(newRecord, true); err != nil {
			return err
		}
		return nil
	})
}

func editProject(t *core.Track, name string) error {
	project, err := t.LoadProjectByName(name)
	if err != nil {
		return err
	}

	return edit(t, &project, func(b []byte) error {
		if err := json.Unmarshal(b, &project); err != nil {
			return err
		}

		if project.Name != name {
			return fmt.Errorf("can't change project name")
		}

		if err = t.SaveProject(project, true); err != nil {
			return err
		}
		return nil
	})
}

func edit(t *core.Track, obj any, fn func(b []byte) error) error {
	file, err := os.CreateTemp("", "track-*.json")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())

	bytes, err := json.MarshalIndent(obj, "", "\t")
	if err != nil {
		return err
	}

	_, err = file.Write(bytes)
	if err != nil {
		return err
	}
	file.Close()

	err = fs.EditFile(file.Name())
	if err != nil {
		return err
	}

	content, err := ioutil.ReadFile(file.Name())
	if err != nil {
		return err
	}

	if len(content) == 0 {
		return ErrUserAbort
	}

	if err := fn(content); err != nil {
		return err
	}

	return nil
}
