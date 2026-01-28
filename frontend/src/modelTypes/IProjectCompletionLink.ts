import type {IAbstract} from './IAbstract'
import type {ILabel} from './ILabel'
import type {IProject} from './IProject'
import type {ITask} from './ITask'
import type {IUser} from './IUser'

export interface IProjectCompletionLink extends IAbstract {
	id: number
	childProjectId: IProject['id']
	parentTaskId: ITask['id']
	labelId: ILabel['id']
	createdBy: IUser | null
	created: Date
	updated: Date
}
