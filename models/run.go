package models

import (
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/lucblassel/training-tracker/config"
	"gorm.io/gorm"
)

type Run struct {
	// Gorm stuff
	ID        uint           `json:"ID" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"CreatedAt"`
	UpdatedAt time.Time      `json:"UpdatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// Data stuff
	Desc       string    `json:"description"`
	Slug       string    `json:"slug" gorm:"index;unique"`
	Path       string    `json:"path"`
	Remote     string    `json:"remote"`
	Finished   bool      `json:"finished"`
	Hidden     bool      `json:"hidden"`
	Successful bool      `json:"successful"`
	LastPulled time.Time `json:"last_pulled"`
	Tags       []*Tag    `json:"tags" gorm:"many2many:run_tags;"`
}

type Tag struct {
	gorm.Model
	Name string `json:"name"`
	Runs []*Run `json:"runs" gorm:"many2many:run_tags;"`
}

type Trace struct {
	X    []int     `json:"x"`
	Y    []float64 `json:"y"`
	Name string    `json:"name"`
	Mode string    `json:"mode"`
	Type string    `json:"type"`
}

func addFloat(s string, array *[]float64) error {
	if strings.ToUpper(s) == "NAN" {
		*array = append(*array, 0)
		return nil
	}

	num, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return err
	}

	*array = append(*array, num)

	return nil
}

func (run *Run) Pull() error {
	cmd := exec.Command("rsync", run.Remote, run.GetFilename())
	log.Println(cmd)
	_, err := cmd.Output()
	if err != nil {
		return err
	}

	stats, _ := os.Stat(run.GetFilename())
	run.LastPulled = stats.ModTime()

	return nil
}

func (run *Run) Init() error {
	_, err := os.Stat(run.GetFilename())
	if err == nil {
		return nil
	}

	if errors.Is(err, os.ErrNotExist) {
		return run.Pull()
	}

	return err
}

func (run *Run) GetFilename() string {
	return path.Join(config.GetDataPath(), fmt.Sprintf("%s.csv", run.Slug))
}

func (run *Run) GetTraces() (map[string]Trace, error) {
	fd, err := os.Open(run.GetFilename())
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	// Read CSV
	reader := csv.NewReader(fd)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	// Populate
	// [timestamp step train_loss val_loss val_MAE val_MRE lr]
	steps := make([]int, 0)
	train_loss, val_loss, lr := make([]float64, 0), make([]float64, 0), make([]float64, 0)

	step_idx, lr_idx, train_idx, val_idx := -1, -1, -1, -1

	for i, record := range records {
		// Skip Header
		if i == 0 {
			for idx, header := range record {
				if header == "step" {
					step_idx = idx
				} else if header == "train_loss" {
					train_idx = idx
				} else if header == "val_loss" {
					val_idx = idx
				} else if header == "lr" {
					lr_idx = idx
				}
			}

			if step_idx == -1 || lr_idx == -1 || train_idx == -1 || val_idx == -1 {
				return nil, errors.New("One header of ['train_loss', 'val_loss', 'step', 'lr'] is missing")
			}

			continue
		}

		step, err := strconv.Atoi(record[step_idx])
		if err != nil {
			return nil, err
		}
		steps = append(steps, step)

		if err := addFloat(record[train_idx], &train_loss); err != nil {
			return nil, err
		}
		if err := addFloat(record[val_idx], &val_loss); err != nil {
			return nil, err
		}
		if err := addFloat(record[lr_idx], &lr); err != nil {
			return nil, err
		}
	}

	traces := map[string]Trace{
		"train_loss": {X: steps, Y: train_loss, Name: "train_loss", Mode: "lines", Type: "scatter"},
		"val_loss":   {X: steps, Y: val_loss, Name: "val_loss", Mode: "lines", Type: "scatter"},
		"lr":         {X: steps, Y: lr, Name: "lr", Mode: "lines", Type: "scatter"},
	}

	return traces, nil
}
