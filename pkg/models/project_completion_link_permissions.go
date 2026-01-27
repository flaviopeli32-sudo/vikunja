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
	"code.vikunja.io/api/pkg/web"
	"xorm.io/xorm"
)

// CanCreate checks if a user can create a project completion link.
func (pcl *ProjectCompletionLink) CanCreate(s *xorm.Session, a web.Auth) (bool, error) {
	return pcl.hasWriteAccess(s, a)
}

// CanRead checks if a user can read a project completion link.
func (pcl *ProjectCompletionLink) CanRead(s *xorm.Session, a web.Auth) (bool, int, error) {
	canWrite, err := pcl.hasWriteAccess(s, a)
	if err != nil || !canWrite {
		return canWrite, 0, err
	}
	return true, int(PermissionWrite), nil
}

func (pcl *ProjectCompletionLink) hasWriteAccess(s *xorm.Session, a web.Auth) (bool, error) {
	if pcl.ChildProjectID == 0 || pcl.ParentTaskID == 0 {
		if pcl.ID == 0 {
			return false, ErrInvalidModel{Message: "child_project_id and parent_task_id are required"}
		}

		lookup := &ProjectCompletionLink{}
		exists, err := s.ID(pcl.ID).Get(lookup)
		if err != nil {
			return false, err
		}
		if !exists {
			return false, ErrProjectCompletionLinkDoesNotExist{ID: pcl.ID}
		}

		pcl.ChildProjectID = lookup.ChildProjectID
		pcl.ParentTaskID = lookup.ParentTaskID
	}

	childProject := &Project{ID: pcl.ChildProjectID}
	canWrite, err := childProject.CanWrite(s, a)
	if err != nil || !canWrite {
		return canWrite, err
	}

	parentProjectID, err := getProjectIDByTaskID(s, pcl.ParentTaskID)
	if err != nil {
		return false, err
	}

	parentProject := &Project{ID: parentProjectID}
	return parentProject.CanWrite(s, a)
}
