// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"time"

	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"
	"xorm.io/xorm"
)

// ProjectCompletionLink represents a relation between a child project and a parent task.
type ProjectCompletionLink struct {
	// The unique, numeric id of this relation.
	ID int64 `xorm:"bigint autoincr not null unique pk" json:"id" param:"link"`
	// The project that should be marked as complete.
	ChildProjectID int64 `xorm:"bigint not null index" json:"child_project_id" query:"child_project_id"`
	// The task that represents the completion for the child project.
	ParentTaskID int64 `xorm:"bigint not null index" json:"parent_task_id"`
	// The label that should be applied once the child project is completed.
	LabelID int64 `xorm:"bigint null" json:"label_id"`

	CreatedByID int64      `xorm:"bigint not null" json:"created_by_id"`
	CreatedBy   *user.User `xorm:"-" json:"created_by,omitempty"`

	// A timestamp when this relation was created. You cannot change this value.
	Created time.Time `xorm:"created not null" json:"created"`
	// A timestamp when this relation was updated. You cannot change this value.
	Updated time.Time `xorm:"updated not null" json:"updated"`

	web.CRUDable    `xorm:"-" json:"-"`
	web.Permissions `xorm:"-" json:"-"`
}

// TableName makes a pretty table name
func (*ProjectCompletionLink) TableName() string {
	return "project_completion_links"
}

// Create creates a project completion link.
// @Summary Create a project completion link
// @Description Creates a relation between a child project and a parent task. The user needs write permissions on both projects.
// @tags project
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param link body models.ProjectCompletionLink true "The project completion link object"
// @Success 201 {object} models.ProjectCompletionLink "The created project completion link."
// @Failure 400 {object} web.HTTPError "Invalid project completion link object provided."
// @Failure 403 {object} web.HTTPError "Not allowed to create the project completion link."
// @Failure 500 {object} models.Message "Internal error"
// @Router /project-completion-links [put]
func (pcl *ProjectCompletionLink) Create(s *xorm.Session, a web.Auth) (err error) {
	if pcl.ChildProjectID == 0 || pcl.ParentTaskID == 0 {
		return ErrInvalidModel{Message: "child_project_id and parent_task_id are required"}
	}

	pcl.CreatedBy, err = GetUserOrLinkShareUser(s, a)
	if err != nil {
		return err
	}
	pcl.CreatedByID = pcl.CreatedBy.ID
	pcl.ID = 0

	_, err = s.Insert(pcl)
	return err
}

// ReadOne returns a project completion link.
// @Summary Get one project completion link
// @Description Returns a project completion link by ID.
// @tags project
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param link path int true "Project completion link ID"
// @Success 200 {object} models.ProjectCompletionLink "The project completion link."
// @Failure 404 {object} web.HTTPError "The project completion link was not found."
// @Failure 500 {object} models.Message "Internal error"
// @Router /project-completion-links/{link} [get]
func (pcl *ProjectCompletionLink) ReadOne(s *xorm.Session, _ web.Auth) (err error) {
	exists, err := s.ID(pcl.ID).Get(pcl)
	if err != nil {
		return err
	}
	if !exists {
		return ErrProjectCompletionLinkDoesNotExist{ID: pcl.ID}
	}

	pcl.CreatedBy, err = user.GetUserByID(s, pcl.CreatedByID)
	if err != nil && !user.IsErrUserDoesNotExist(err) {
		return err
	}

	return nil
}

// ReadAll gets all project completion links for a child project.
// @Summary Get project completion links for a child project
// @Description Returns all project completion links for a child project.
// @tags project
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param child_project_id query int true "Child Project ID"
// @Success 200 {array} models.ProjectCompletionLink "The project completion links"
// @Failure 400 {object} web.HTTPError "Invalid project completion link filter provided."
// @Failure 403 {object} web.HTTPError "Not allowed to read the project completion links."
// @Failure 500 {object} models.Message "Internal error"
// @Router /project-completion-links [get]
func (pcl *ProjectCompletionLink) ReadAll(s *xorm.Session, a web.Auth, _ string, _ int, _ int) (result interface{}, resultCount int, numberOfTotalItems int64, err error) {
	if pcl.ChildProjectID == 0 {
		return nil, 0, 0, ErrInvalidModel{Message: "child_project_id is required"}
	}

	childProject := &Project{ID: pcl.ChildProjectID}
	canWrite, err := childProject.CanWrite(s, a)
	if err != nil {
		return nil, 0, 0, err
	}
	if !canWrite {
		return nil, 0, 0, ErrGenericForbidden{}
	}

	links := []*ProjectCompletionLink{}
	err = s.Where("child_project_id = ?", pcl.ChildProjectID).Find(&links)
	if err != nil {
		return nil, 0, 0, err
	}

	filtered := make([]*ProjectCompletionLink, 0, len(links))
	for _, link := range links {
		parentProjectID, err := getProjectIDByTaskID(s, link.ParentTaskID)
		if err != nil {
			return nil, 0, 0, err
		}

		parentProject := &Project{ID: parentProjectID}
		canWrite, err = parentProject.CanWrite(s, a)
		if err != nil {
			return nil, 0, 0, err
		}
		if !canWrite {
			continue
		}

		filtered = append(filtered, link)
	}

	return filtered, len(filtered), int64(len(filtered)), nil
}

func getProjectIDByTaskID(s *xorm.Session, taskID int64) (int64, error) {
	task, err := GetTaskByIDSimple(s, taskID)
	if err != nil {
		return 0, err
	}
	return task.ProjectID, nil
}
