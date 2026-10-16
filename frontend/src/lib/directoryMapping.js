const ASSET_ENTITY = 'asset';
const SHOT_ENTITY = 'shot';
const TASK_ENTITY = 'task';

const ASSET_PLACEHOLDER = '<Asset>';
const SHOT_PLACEHOLDER = '<Shot>';

export const inferDirectoryTemplateEntity = (id, template = '') => {
  const value = `${id || ''} ${template || ''}`.toLowerCase();
  if (value.includes(SHOT_PLACEHOLDER.toLowerCase()) || value.includes(SHOT_ENTITY)) return SHOT_ENTITY;
  if (value.includes(ASSET_PLACEHOLDER.toLowerCase()) || value.includes(ASSET_ENTITY)) return ASSET_ENTITY;
  return '';
};

export const filterTaskTypesForTemplate = (template, taskTypes) => {
  const entity = inferDirectoryTemplateEntity(template.id, template.template);
  if (!entity) return taskTypes;

  const hasEntityScopedTypes = taskTypes.some(taskType => {
    const scope = String(taskType.for_entity || '').toLowerCase();
    return scope && scope !== TASK_ENTITY;
  });
  if (!hasEntityScopedTypes) return taskTypes;

  return taskTypes.filter(taskType => {
    const scope = String(taskType.for_entity || '').toLowerCase();
    return !scope || scope === TASK_ENTITY || scope === entity;
  });
};

export const getDefaultTaskOutput = (template, taskType) => {
  const entity = inferDirectoryTemplateEntity(template.id, template.template);
  const subjectPlaceholder = entity === SHOT_ENTITY ? SHOT_PLACEHOLDER : ASSET_PLACEHOLDER;
  return `${subjectPlaceholder}-${String(taskType || '').toLowerCase()}`;
};

export const fillDefaultTaskOutputs = (template, taskTypes) => {
  const taskOutputs = { ...(template.task_outputs || {}) };

  for (const taskType of taskTypes) {
    const taskTypeName = taskType.name;
    const isEnabled = template.task_output_enabled?.[taskTypeName] !== false;
    const hasOutput = String(taskOutputs[taskTypeName] || '').trim() !== '';
    if (!taskTypeName || !isEnabled || hasOutput) continue;

    taskOutputs[taskTypeName] = getDefaultTaskOutput(template, taskTypeName);
  }

  return taskOutputs;
};

export const getDirectoryTemplatePreviewValues = (template) => {
  const entity = inferDirectoryTemplateEntity(template.id, template.template);
  const isShotTemplate = entity === SHOT_ENTITY;
  const taskType = isShotTemplate ? 'animation' : 'modeling';
  const subject = isShotTemplate ? 'shot010' : 'hero';

  return {
    Episode: 'ep01',
    Sequence: 'seq001',
    Shot: 'shot010',
    CollectionType: 'character',
    Category: 'heroes',
    Asset: 'hero',
    AssetType: taskType,
    OutputName: `${subject}-${taskType}`,
    TemplateExtension: '.ext',
  };
};
