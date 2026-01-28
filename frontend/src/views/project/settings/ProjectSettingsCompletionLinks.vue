<script setup lang="ts">
import {computed, ref, shallowReactive, watchEffect} from 'vue'
import {useRoute} from 'vue-router'
import {useI18n} from 'vue-i18n'
import {useTitle} from '@vueuse/core'

import ProjectService from '@/services/project'
import ProjectModel from '@/models/project'
import ProjectCompletionLinkModel from '@/models/projectCompletionLink'
import ProjectCompletionLinkService from '@/services/projectCompletionLink'
import TaskService from '@/services/task'
import TaskModel from '@/models/task'
import LabelModel from '@/models/label'
import type {IProject} from '@/modelTypes/IProject'
import type {ITask} from '@/modelTypes/ITask'
import type {ILabel} from '@/modelTypes/ILabel'
import type {IProjectCompletionLink} from '@/modelTypes/IProjectCompletionLink'

import CreateEdit from '@/components/misc/CreateEdit.vue'
import FormField from '@/components/input/FormField.vue'
import Multiselect from '@/components/input/Multiselect.vue'
import XButton from '@/components/input/Button.vue'
import LabelTag from '@/components/tasks/partials/Label.vue'

import {useBaseStore} from '@/stores/base'
import {useLabelStore} from '@/stores/labels'
import {useProjectStore} from '@/stores/projects'
import {error, success} from '@/message'
import {getRandomColorHex} from '@/helpers/color/randomColor'
import {formatDateShort} from '@/helpers/time/formatDate'

type TaskSearchResult = ITask & {differentProject?: string | null}

defineOptions({name: 'ProjectSettingCompletionLinks'})

const {t} = useI18n({useScope: 'global'})

const project = ref<IProject>()
const links = ref<IProjectCompletionLink[]>([])
const linksLoaded = ref(false)

const projectStore = useProjectStore()
const labelStore = useLabelStore()

const projectCompletionLinkService = shallowReactive(new ProjectCompletionLinkService())
const taskService = shallowReactive(new TaskService())

const selectedTask = ref<ITask | null>(null)
const selectedLabel = ref<ILabel | null>(null)
const isSaving = ref(false)

const taskQuery = ref('')
const foundTasks = ref<ITask[]>([])
const labelQuery = ref('')

const tasksById = ref<Record<number, ITask>>({})

const route = useRoute()
const projectId = computed(() => route.params.projectId !== undefined
	? parseInt(route.params.projectId as string)
	: undefined,
)

const title = computed(() => project.value?.title
	? t('project.completionLinks.titleWithProject', {project: project.value.title})
	: t('project.completionLinks.title'),
)
useTitle(title)

watchEffect(() => projectId.value !== undefined && loadProject(projectId.value))

async function loadProject(projectId: number) {
	const projectService = new ProjectService()
	const newProject = await projectService.get(new ProjectModel({id: projectId}))
	await useBaseStore().handleSetCurrentProject({project: newProject})
	project.value = newProject
	await labelStore.loadAllLabels()
	await loadCompletionLinks()
}

async function loadCompletionLinks() {
	if (!project.value) {
		return
	}

	linksLoaded.value = false
	links.value = await projectCompletionLinkService.getAll({childProjectId: project.value.id})
	await ensureTasksLoaded(links.value)
	linksLoaded.value = true
}

function mapTasks(tasks: ITask[]): TaskSearchResult[] {
	return tasks.map(task => {
		const projectInfo = projectStore.projects[task.projectId]

		return {
			...task,
			differentProject: (projectInfo && task.projectId !== projectId.value && projectInfo.title) || null,
		}
	})
}

const mappedFoundTasks = computed(() => mapTasks(foundTasks.value))

async function findTasks(newQuery: string) {
	taskQuery.value = newQuery
	const result = await taskService.getAll({}, {
		s: newQuery,
		sort_by: 'done',
	})

	foundTasks.value = result
}

async function ensureTasksLoaded(currentLinks: IProjectCompletionLink[]) {
	const missingIds = currentLinks
		.map(link => link.parentTaskId)
		.filter(taskId => typeof tasksById.value[taskId] === 'undefined')

	if (missingIds.length === 0) {
		return
	}

	const tasks = await Promise.all(
		missingIds.map(taskId => taskService.get(new TaskModel({id: taskId}))),
	)

	tasks.forEach(task => {
		tasksById.value[task.id] = task
	})
}

const linkTasks = computed<Record<number, TaskSearchResult | null>>(() => {
	return links.value.reduce((acc, link) => {
		const task = tasksById.value[link.parentTaskId]
		acc[link.id] = task ? mapTasks([task])[0] : null
		return acc
	}, {} as Record<number, TaskSearchResult | null>)
})

const foundLabels = computed(() => labelStore.filterLabelsByQuery([], labelQuery.value))

function findLabel(newQuery: string) {
	labelQuery.value = newQuery
}

async function createLabel(title: string) {
	const newLabel = await labelStore.createLabel(new LabelModel({
		title,
		hexColor: getRandomColorHex(),
	}))
	selectedLabel.value = newLabel
	success({message: t('task.label.createSuccess')})
}

async function createLink() {
	if (isSaving.value) {
		return
	}

	if (!selectedTask.value?.id) {
		error({message: t('project.completionLinks.taskRequired')})
		return
	}

	if (!project.value) {
		return
	}

	isSaving.value = true

	try {
		const created = await projectCompletionLinkService.create(new ProjectCompletionLinkModel({
			childProjectId: project.value.id,
			parentTaskId: selectedTask.value.id,
			labelId: selectedLabel.value?.id ?? 0,
		}))
		links.value.push(created)
		await ensureTasksLoaded([created])
		selectedTask.value = null
		selectedLabel.value = null
		success({message: t('project.completionLinks.createSuccess')})
	} finally {
		isSaving.value = false
	}
}
</script>

<template>
	<CreateEdit
		:title="title"
		:has-primary-action="false"
		:wide="true"
	>
		<p class="mbs-4">
			{{ $t('project.completionLinks.description') }}
		</p>

		<div class="box mbe-4">
			<FormField :label="$t('project.completionLinks.parentTask')">
				<Multiselect
					v-model="selectedTask"
					:loading="taskService.loading"
					:placeholder="$t('project.completionLinks.parentTaskPlaceholder')"
					:search-results="mappedFoundTasks"
					label="title"
					@search="findTasks"
				>
					<template #searchResult="{option: task}">
						<span
							v-if="typeof task !== 'string'"
							class="search-result"
							:class="{'is-strikethrough': task.done}"
						>
							<span
								v-if="task.differentProject"
								v-tooltip="$t('task.relation.differentProject')"
								class="different-project"
							>
								{{ task.differentProject }} &gt;
							</span>
							{{ task.title }}
						</span>
						<span
							v-else
							class="search-result"
						>
							{{ task }}
						</span>
					</template>
				</Multiselect>
			</FormField>

			<FormField :label="$t('project.completionLinks.label')">
				<Multiselect
					v-model="selectedLabel"
					:loading="labelStore.isLoading"
					:placeholder="$t('project.completionLinks.labelPlaceholder')"
					:search-results="foundLabels"
					label="title"
					:creatable="true"
					:create-placeholder="$t('task.label.createPlaceholder')"
					@search="findLabel"
					@create="createLabel"
				>
					<template #searchResult="{option}">
						<LabelTag
							v-if="typeof option !== 'string'"
							:label="option"
							class="search-result"
						/>
						<span
							v-else
							class="tag search-result"
						>
							<span>{{ option }}</span>
						</span>
					</template>
				</Multiselect>
				<p class="help">
					{{ $t('project.completionLinks.labelHint') }}
				</p>
			</FormField>

			<XButton
				:loading="isSaving"
				:disabled="!selectedTask"
				icon="plus"
				@click="createLink"
			>
				{{ $t('project.completionLinks.create') }}
			</XButton>
		</div>

		<table
			v-if="links.length > 0"
			class="table is-fullwidth is-striped is-hoverable"
		>
			<thead>
				<tr>
					<th>{{ $t('project.completionLinks.columns.task') }}</th>
					<th>{{ $t('project.completionLinks.columns.project') }}</th>
					<th>{{ $t('project.completionLinks.columns.label') }}</th>
					<th>{{ $t('project.completionLinks.columns.created') }}</th>
				</tr>
			</thead>
			<tbody>
				<tr
					v-for="link in links"
					:key="link.id"
				>
					<template v-if="linkTasks[link.id]">
						<td>
							<RouterLink
								:to="{ name: 'task.detail', params: { id: link.parentTaskId } }"
							>
								{{ linkTasks[link.id]?.title }}
							</RouterLink>
						</td>
						<td>
							{{ projectStore.projects[linkTasks[link.id]?.projectId || 0]?.title || '-' }}
						</td>
						<td>
							<LabelTag
								v-if="link.labelId && labelStore.getLabelById(link.labelId)"
								:label="labelStore.getLabelById(link.labelId)"
							/>
							<span v-else>-</span>
						</td>
						<td>
							{{ formatDateShort(link.created) }}
						</td>
					</template>
					<template v-else>
						<td colspan="4">
							{{ $t('project.completionLinks.loadingTask') }}
						</td>
					</template>
				</tr>
			</tbody>
		</table>

		<p
			v-else-if="linksLoaded"
			class="has-text-grey"
		>
			{{ $t('project.completionLinks.noneYet') }}
		</p>
	</CreateEdit>
</template>

<style scoped lang="scss">
.search-result {
	display: inline-flex;
	align-items: center;
	gap: 0.25rem;
}

.different-project {
	color: var(--grey-600);
	font-size: 0.85em;
	margin-inline-end: 0.35rem;
}
</style>
