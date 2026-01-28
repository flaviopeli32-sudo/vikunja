import AbstractService from './abstractService'
import ProjectCompletionLinkModel from '@/models/projectCompletionLink'
import type {IProjectCompletionLink} from '@/modelTypes/IProjectCompletionLink'

export default class ProjectCompletionLinkService extends AbstractService<IProjectCompletionLink> {
	constructor() {
		super({
			create: '/project-completion-links',
			getAll: '/project-completion-links',
			get: '/project-completion-links/{id}',
		})
	}

	modelFactory(data) {
		return new ProjectCompletionLinkModel(data)
	}
}
