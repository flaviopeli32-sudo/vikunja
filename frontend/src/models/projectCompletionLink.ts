import AbstractModel from '@/models/abstractModel'
import UserModel from '@/models/user'
import type {IProjectCompletionLink} from '@/modelTypes/IProjectCompletionLink'

export default class ProjectCompletionLinkModel extends AbstractModel<IProjectCompletionLink> implements IProjectCompletionLink {
	id = 0
	childProjectId = 0
	parentTaskId = 0
	labelId = 0
	createdBy = null
	created: Date
	updated: Date

	constructor(data: Partial<IProjectCompletionLink> = {}) {
		super()
		this.assignData(data)

		this.createdBy = this.createdBy ? new UserModel(this.createdBy) : null
		this.created = new Date(this.created)
		this.updated = new Date(this.updated)
	}
}
