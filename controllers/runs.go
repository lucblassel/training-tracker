package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lucblassel/training-tracker/generator"
	"github.com/lucblassel/training-tracker/models"
)

func GetAllRuns(c *gin.Context) {
	var runs []models.Run
	models.DB.Order("created_at desc, updated_at desc, id desc").Find(&runs)

	c.HTML(http.StatusOK, "index.html", gin.H{"title": "All training runs", "runs": runs})
}

func GetRuns(c *gin.Context) {
	var runs []models.Run
	models.DB.Where(map[string]interface{}{"Hidden": false}).Find(&runs)

	c.HTML(http.StatusOK, "index.html", gin.H{"title": "Training runs", "runs": runs})
}

func GetFinishedRuns(c *gin.Context) {
	var runs []models.Run
	models.DB.Where(models.Run{Finished: true}).Find(&runs)
	c.HTML(http.StatusOK, "index.html", gin.H{"title": "Finished runs", "runs": runs})
}

func GetRunningRuns(c *gin.Context) {
	var runs []models.Run
	models.DB.Where(map[string]interface{}{"Finished": false}).Find(&runs)
	c.HTML(http.StatusOK, "index.html", gin.H{"title": "Finished runs", "runs": runs})
}

func GetRun(c *gin.Context) {
	var run models.Run
	if result := models.DB.Where("slug = ?", c.Param("slug")).First(&run); result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}

	run.Init()

	p, _ := json.MarshalIndent(run, "", "\t")

	traces, err := run.GetTraces()
	var currentStep int
	if err != nil {
		log.Printf("Error loading CSV: %s\n", err)
	} else {
		currentStep = traces["lr"].X[len(traces["lr"].X)-1]
	}

	t, _ := json.MarshalIndent(traces, "", "\t")

	c.HTML(
		http.StatusOK,
		"run.html",
		gin.H{"title": run.Slug, "Run": run, "PrettyJson": fmt.Sprintf("%s\n\n%s", p, t), "Plotly": true, "Traces": traces, "error": err, "currentStep": currentStep},
	)
}

func PullRun(c *gin.Context) {
	var run models.Run
	if result := models.DB.Where("slug = ?", c.Param("slug")).First(&run); result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}
	if !run.Finished {
		if err := run.Pull(); err != nil {
			log.Println(err)
			c.Redirect(http.StatusSeeOther, fmt.Sprintf("/run/%s", run.Slug))
		}
		models.DB.Save(&run)
	}

	c.Redirect(http.StatusFound, fmt.Sprintf("/run/%s", run.Slug))
}

func ToggleRunFinished(c *gin.Context) {
	var run models.Run
	if result := models.DB.Where("slug = ?", c.Param("slug")).First(&run); result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}

	new_status := !run.Finished
	if result := models.DB.Model(&run).Where("id = ?", run.ID).Update("finished", new_status); result.Error != nil {
		log.Println(result.Error)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error changing finished status"})
		return
	}

	c.Redirect(http.StatusFound, fmt.Sprintf("/run/%s", run.Slug))
}

func MarkRunFailed(c *gin.Context) {
	var run models.Run
	if result := models.DB.Where("slug = ?", c.Param("slug")).First(&run); result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}

	if result := models.DB.Model(&run).Where("id = ?", run.ID).Updates(map[string]interface{}{"finished": true, "successful": false}); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error setting run as failed"})
		return
	}

	c.Redirect(http.StatusFound, fmt.Sprintf("/run/%s", run.Slug))
}

func MarkRunSuccessful(c *gin.Context) {
	var run models.Run
	if result := models.DB.Where("slug = ?", c.Param("slug")).First(&run); result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}

	if result := models.DB.Model(&run).Where("id = ?", run.ID).Updates(map[string]interface{}{"finished": true, "successful": true}); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error setting run as successful"})
		return
	}

	c.Redirect(http.StatusFound, fmt.Sprintf("/run/%s", run.Slug))
}

func EditRun(c *gin.Context) {
	var run models.Run
	if result := models.DB.Where("slug = ?", c.Param("slug")).First(&run); result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}

	id := c.Query("id")

	p, _ := json.MarshalIndent(run, "", "\t")
	c.HTML(
		http.StatusOK,
		"edit.html",
		gin.H{"title": run.Slug, "Run": run, "PrettyJson": fmt.Sprintf("%s\n", p), "id": id},
	)
}

func SaveRun(c *gin.Context) {
	var form CreateRunInput
	c.ShouldBind(&form)

	id := c.Query("id")
	var run models.Run
	if id != "" {
		if result := models.DB.Where("id = ?", id).First(&run); result.Error != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found"})
			return
		}
	}

	run.Slug = form.Slug
	run.Desc = form.Desc
	run.Remote = form.Remote

	if result := models.DB.Save(&run); result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error while updating record"})
		return
	}

	c.Redirect(http.StatusFound, fmt.Sprintf("/run/%s", run.Slug))
}

type CreateRunInput struct {
	Desc     string `json:"description" binding:"required" form:"run_description"`
	Slug     string `json:"slug" binding:"required" form:"run_name"`
	Path     string `json:"path"`
	Remote   string `json:"remote" binding:"required" form:"run_remote"`
	Finished bool   `json:"finished" form:"run_finished"`
	Hidden   bool   `json:"hidden" form:"run_hidden"`
}

func CreateRun(c *gin.Context) {
	slug := generator.GenerateID()
	c.HTML(
		http.StatusOK,
		"edit.html",
		gin.H{"title": slug, "Run": models.Run{Slug: slug}},
	)
}

func DuplicateRun(c *gin.Context) {
	var run models.Run
	if result := models.DB.Where("slug = ?", c.Param("slug")).First(&run); result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}

	newRun := models.Run{
		Slug:   generator.GenerateID(),
		Desc:   run.Desc,
		Remote: run.Remote,
	}

	c.HTML(
		http.StatusOK,
		"edit.html",
		gin.H{"title": newRun.Slug, "Run": newRun},
	)
}

func DeleteRun(c *gin.Context) {
	var run models.Run
	if result := models.DB.Where("slug = ?", c.Param("slug")).First(&run); result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}

	if result := models.DB.Delete(&run); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting record"})
		return
	}

	c.Redirect(http.StatusFound, "/")
}

func UpdateAll(c *gin.Context) {
	var runs []models.Run
	if results := models.DB.Where("finished = ?", false).Find(&runs); results.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": results.Error})
	}

	for _, run := range runs {
		if err := run.Pull(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		}
	}

	// Update last pulled
	if len(runs) > 0 {
		models.DB.Save(&runs)
	}

	c.Redirect(http.StatusFound, "/")
}

func GetTags(c *gin.Context) {
	var tags []models.Tag
	if result := models.DB.Find(&tags); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch tags from DB"})
		return
	}

	c.HTML(http.StatusOK, "tags.html", gin.H{"Tags": tags})
}

func SaveTag(c *gin.Context) {
	var form CreateTagInput
	c.ShouldBind(&form)

	id := c.Query("id")
	var tag models.Tag
	if id != "" {
		if result := models.DB.Where("id = ?", id).First(&tag); result.Error != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Tag not found"})
			return
		}
	}

	tag.Name = form.Name
	tag.Color = form.Color

	if result := models.DB.Save(&tag); result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error whil updating tag"})
		return
	}

	c.Redirect(http.StatusFound, "/tags")
}

type CreateTagInput struct {
	Name  string `json:"name" binding:"required" form:"tag_name"`
	Color string `json:"color" bondong:"required" form:"tag_color"`
}

func CreateTag(c *gin.Context) {
	c.HTML(
		http.StatusOK,
		"edit_tag.html",
		gin.H{"title": "New Tag", "Tag": models.Tag{}, "Colors": models.TAGCOLORS},
	)
}
